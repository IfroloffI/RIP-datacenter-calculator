package my.froloff.calc.dto;

import lombok.AllArgsConstructor;
import lombok.Data;

@Data
@AllArgsConstructor
public class ResultForWebHookRequest {
    private Long calculation_id;
    private Integer total_power;
    private Boolean success;
}
