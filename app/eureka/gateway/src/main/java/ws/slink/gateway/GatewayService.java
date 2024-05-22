package ws.slink.gateway;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.cloud.netflix.eureka.EnableEurekaClient;

@EnableEurekaClient
@SpringBootApplication
public class GatewayService {
    public static void main(String[] args) {
        SpringApplication.run(GatewayService.class, args);
    }
}
