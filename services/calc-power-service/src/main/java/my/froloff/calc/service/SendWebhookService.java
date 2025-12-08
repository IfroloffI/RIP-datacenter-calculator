package my.froloff.calc.service;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import my.froloff.calc.dto.ResultForWebHookRequest;
import org.slf4j.MDC;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestClient;

@Slf4j
@Service
@RequiredArgsConstructor
public class SendWebhookService {

    private final RestClient.Builder restClientBuilder;

    @Value("${app.feature.async-token}")
    private String token;

    @Value("${app.feature.webhook-base-url}")
    private String baseUrlForSendResult;

    @Value("${app.feature.webhook-endpoint}")
    private String endpointForSendResult;

    private static final int MAX_RETRIES = 3;

    @Async("calculationWebHook")
    public void sendWebhookWithRetry(ResultForWebHookRequest result, String traceId) {
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
                            .uri(endpointForSendResult)
                            .header("X-Async-Token", token)
                            .header("X-Trace-Id", traceId != null ? traceId : "")
                            .body(result)
                            .retrieve()
                            .toBodilessEntity();

                    log.info("Webhook sent successfully on attempt {}, traceId={}", attempt, traceId);
                    return;

                } catch (org.springframework.web.client.HttpClientErrorException ex) {
                    log.warn("Webhook client error ({}) for calcId={}, attempt {}, traceId={}",
                            ex.getStatusCode(), result.getCalculation_id(), attempt, traceId);
                    return;

                } catch (org.springframework.web.client.HttpServerErrorException ex) {
                    log.warn("Webhook server error ({}) for calcId={}, attempt {}, traceId={}",
                            ex.getStatusCode(), result.getCalculation_id(), attempt, traceId);

                    if (attempt >= MAX_RETRIES) {
                        log.warn("All webhook attempts exhausted (5xx) for calcId={}, traceId={}",
                                result.getCalculation_id(), traceId);
                        return;
                    }

                    try {
                        Thread.sleep(500L * attempt);
                    } catch (InterruptedException ie) {
                        Thread.currentThread().interrupt();
                        log.warn("Webhook retry interrupted, traceId={}", traceId);
                        return;
                    }

                } catch (Exception e) {
                    log.warn("Webhook failed (attempt {}, {}), calcId={}, traceId={}",
                            attempt, e.getClass().getSimpleName(), result.getCalculation_id(), traceId);

                    if (attempt >= MAX_RETRIES) {
                        log.warn("All webhook attempts exhausted (exception) for calcId={}, traceId={}",
                                result.getCalculation_id(), traceId);
                        return;
                    }

                    try {
                        Thread.sleep(1000L * attempt);
                    } catch (InterruptedException ie) {
                        Thread.currentThread().interrupt();
                        log.warn("Webhook retry interrupted, traceId={}", traceId);
                        return;
                    }
                }
            }
        } finally {
            MDC.remove("traceId");
        }
    }
}
