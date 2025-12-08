package my.froloff.calc.controller.http;

import jakarta.validation.Valid;
import my.froloff.calc.dto.AcceptedResponse;
import my.froloff.calc.dto.CalculationRequest;
import my.froloff.calc.dto.ErrorResponse;
import my.froloff.calc.dto.Response;
import my.froloff.calc.service.CalculationService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.slf4j.MDC;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@Slf4j
@RestController
@RequestMapping("/power-calculations")
@RequiredArgsConstructor
public class CalculationController {

    private final CalculationService calculationService;
    private static final int START_MAX_RETRIES = 3;

    @PostMapping("/calculate")
    public ResponseEntity<Response> startCalculation(@RequestBody @Valid CalculationRequest request) {
        String traceId = MDC.get("traceId");
        log.info("Received async calculation request: {}", request);

        int attempt = 0;
        while (true) {
            attempt++;
            try {
                calculationService.startAsyncCalculation(request, traceId);
                break;
            } catch (Exception e) {
                log.warn("Failed to submit async calculation on attempt {}, request={}", attempt, request, e);
                if (attempt >= START_MAX_RETRIES) {
                    log.error("All attempts to submit async calculation failed, request={}", request, e);
                    return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(new ErrorResponse("Failed to start async calculation"));
                }
                try {
                    Thread.sleep(200L * attempt);
                } catch (InterruptedException ie) {
                    Thread.currentThread().interrupt();
                    return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(new ErrorResponse("Interrupted while starting async calculation"));
                }
            }
        }

        return ResponseEntity.status(HttpStatus.ACCEPTED).body(new AcceptedResponse("Status update initiated"));
    }
}
