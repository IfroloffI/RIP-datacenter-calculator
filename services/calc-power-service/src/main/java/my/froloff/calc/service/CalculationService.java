package my.froloff.calc.service;

import my.froloff.calc.dto.CalculationRequest;
import my.froloff.calc.dto.ResultForWebHookRequest;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.slf4j.MDC;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Service;

import java.util.concurrent.ThreadLocalRandom;

@Slf4j
@Service
@RequiredArgsConstructor
public class CalculationService {

    private final SendWebhookService sendWebhookService;

    @Async("calculationExecutor")
    public void startAsyncCalculation(CalculationRequest request, String traceId) {
        MDC.put("traceId", traceId);

        try {
            long totalPower = request.getDevices().stream().mapToLong(d -> (long) d.getPower_watt() * d.getQuantity()).sum();
            int totalPowerWithPUE = (int) (totalPower * request.getPue());

            ResultForWebHookRequest successResult = new ResultForWebHookRequest(request.getCalculation_id(), totalPowerWithPUE, true);

            long delayMillis = ThreadLocalRandom.current().nextLong(2500, 5001);
            log.info("Simulating long calculation: sleep {} ms, traceId={}", delayMillis, traceId);
            Thread.sleep(delayMillis);

            log.info("Async calculation done, traceId={}, result={}", traceId, successResult);
            sendWebhookService.sendWebhookWithRetry(successResult, traceId);

        } catch (Exception e) {
            log.error("Async calculation failed, traceId={}", traceId, e);
            ResultForWebHookRequest failureResult = new ResultForWebHookRequest(request.getCalculation_id(), 0, false);
            sendWebhookService.sendWebhookWithRetry(failureResult, traceId);
        } finally {
            MDC.remove("traceId");
        }
    }
}
