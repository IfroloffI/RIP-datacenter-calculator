package my.froloff.calc.config;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.scheduling.annotation.EnableAsync;

import java.util.concurrent.Executor;
import java.util.concurrent.Executors;

@Configuration
@EnableAsync
public class AsyncConfig {

    @Bean(name = "calculationExecutor")
    public Executor calculationExecutor() {
        return Executors.newThreadPerTaskExecutor(Thread.ofVirtual().name("calc-", 0).factory());
    }

    @Bean(name = "calculationWebHook")
    public Executor calculationWebHookExecutor() {
        return Executors.newThreadPerTaskExecutor(Thread.ofVirtual().name("webhook-", 0).factory());
    }

}
