package my.froloff.calc.service;

import my.froloff.calc.dto.CalculationRequest;
import my.froloff.calc.dto.ResultForWebHookRequest;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.slf4j.MDC;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestClient;

import java.util.concurrent.ThreadLocalRandom;

@Slf4j
@Service
@RequiredArgsConstructor
public class CalculationService {

    private final RestClient.Builder restClientBuilder;

    @Value("${app.feature.async-token}")
    private String token;

    @Value("${app.feature.webhook-base-url}")
    private String baseUrlForSendResult;

    @Value("${app.feature.webhook-endpoint}")
    private String EndpointForSendResult;

    private static final int MAX_RETRIES = 3;

    @Async("calculationExecutor")
    public void startAsyncCalculation(CalculationRequest request, String traceId) {
        MDC.put("traceId", traceId);

        try {
            long totalPower = request.getDevices().stream().mapToLong(d -> (long) d.getPower_watt() * d.getQuantity()).sum();
            int totalPowerWithPUE = (int) (totalPower * request.getPue());

            ResultForWebHookRequest successResult = new ResultForWebHookRequest(request.getCalculation_id(), totalPowerWithPUE, true);

            // 2.5 - 5 сек иммитированная нагрузка
            long delayMillis = ThreadLocalRandom.current().nextLong(2500, 5001);
            log.info("Simulating long calculation: sleep {} ms, traceId={}", delayMillis, traceId);
            Thread.sleep(delayMillis);

            log.info("Async calculation done, traceId={}, result={}", traceId, successResult);
            sendWebhookWithRetry(successResult, traceId);

        } catch (Exception e) {
            log.error("Async calculation failed, traceId={}", traceId, e);
            ResultForWebHookRequest failureResult = new ResultForWebHookRequest(request.getCalculation_id(), 0, false);
            sendWebhookWithRetry(failureResult, traceId);
        } finally {
            MDC.remove("traceId");
        }
    }

    void sendWebhookWithRetry(ResultForWebHookRequest result, String traceId) {
        MDC.put("traceId", traceId);
        try {
            RestClient client = restClientBuilder.baseUrl(baseUrlForSendResult).build();

            int attempt = 0;
            while (true) {
                attempt++;
                try {
                    log.info("Sending webhook attempt {} for calcId={}, success={}, traceId={}",
                            attempt, result.getCalculation_id(), result.getSuccess(), traceId);

                    client.put()
                            .uri(EndpointForSendResult)
                            .header("X-Async-Token", token)
                            .header("X-Trace-Id", traceId != null ? traceId : "")
                            .body(result)
                            .retrieve()
                            .toBodilessEntity();

                    log.info("Webhook sent successfully on attempt {}, traceId={}", attempt, traceId);
                    return;

                } catch (org.springframework.web.client.HttpClientErrorException ex) {
                    log.error("Webhook failed with client error ({}): calcId={}, traceId={}, error={}",
                            ex.getStatusCode(), result.getCalculation_id(), traceId, ex.getResponseBodyAsString());
                    return;

                } catch (org.springframework.web.client.HttpServerErrorException ex) {
                    log.warn("Webhook attempt {} failed with server error ({}) for calcId={}, traceId={}",
                            attempt, ex.getStatusCode(), result.getCalculation_id(), traceId);

                    if (attempt >= MAX_RETRIES) {
                        log.error("All webhook attempts failed (5xx) for calcId={}, traceId={}",
                                result.getCalculation_id(), traceId);
                        return;
                    }

                    try {
                        Thread.sleep(500L * attempt);
                    } catch (InterruptedException ie) {
                        Thread.currentThread().interrupt();
                        log.error("Webhook retry interrupted, traceId={}", traceId, ie);
                        return;
                    }

                } catch (Exception e) {
                    log.warn("Webhook attempt {} failed with exception for calcId={}, traceId={}",
                            attempt, result.getCalculation_id(), traceId, e);

                    if (attempt >= MAX_RETRIES) {
                        log.error("All webhook attempts failed (exception) for calcId={}, traceId={}",
                                result.getCalculation_id(), traceId);
                        return;
                    }

                    try {
                        Thread.sleep(1000L * attempt);
                    } catch (InterruptedException ie) {
                        Thread.currentThread().interrupt();
                        log.error("Webhook retry interrupted, traceId={}", traceId, ie);
                        return;
                    }
                }
            }
        } finally {
            MDC.remove("traceId");
        }
    }
}
