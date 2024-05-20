package com.rt.solar.cyberrange.discovery;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.config.annotation.web.configuration.EnableWebSecurity;
import org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter;
import org.springframework.security.web.util.matcher.AntPathRequestMatcher;

// https://www.appsdeveloperblog.com/secure-eureka-dashboard-with-spring-security/

@EnableWebSecurity
public class SecurityConfig extends WebSecurityConfigurerAdapter {

    @Value("${spring.security.user.name}")
    private String userName;

    @Value("${spring.security.user.password}")
    private String userPassword;

    @Override
    protected void configure(HttpSecurity http) throws Exception {
        if ("x".equals(userName) && "x".equals(userPassword)) {
            http
                .csrf().disable()
                .httpBasic().disable();
        } else {
            http
                .csrf().disable()
                .authorizeRequests()
                    .requestMatchers(new AntPathRequestMatcher("/actuator/**", "GET")).permitAll()
                    .anyRequest().authenticated()
                .and()
                .httpBasic();
        }
    }
}
