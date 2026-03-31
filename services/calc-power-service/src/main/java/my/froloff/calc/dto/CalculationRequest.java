package my.froloff.calc.dto;

import jakarta.validation.Valid;
import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.NotEmpty;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Positive;
import lombok.Data;

import java.util.List;

@Data
public class CalculationRequest {

    @NotNull
    @Min(1)
    private Long calculation_id;

    @NotEmpty
    @Valid
    private List<DevicePayload> devices;

    @NotNull
    @Positive(message = "PUE must be positive")
    private Double pue;
}
