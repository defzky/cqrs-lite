package com.example.productsearch.nats;

import com.example.productsearch.service.ProductIndexService;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import io.nats.client.Connection;
import io.nats.client.Dispatcher;
import io.nats.client.Nats;
import io.nats.client.Options;
import jakarta.annotation.PostConstruct;
import jakarta.annotation.PreDestroy;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.nio.charset.StandardCharsets;
import java.time.Duration;

/**
 * NATS JetStream consumer untuk menerima CDC events dari Debezium Server.
 * 
 * Flow:
 * 1. Debezium Server tangkap perubahan dari PostgreSQL (CDC)
 * 2. Debezium kirim event ke NATS JetStream
 * 3. Consumer ini terima event dan update search index
 */
@Component
@Slf4j
@RequiredArgsConstructor
public class NatsConsumer {
    
    private final ProductIndexService productIndexService;
    private final ObjectMapper objectMapper;
    
    @Value("${nats.url:nats://localhost:4222}")
    private String natsUrl;
    
    private Connection connection;
    private Dispatcher dispatcher;
    
    @PostConstruct
    public void start() throws Exception {
        log.info("Connecting to NATS at {}", natsUrl);
        
        Options options = new Options.Builder()
                .server(natsUrl)
                .connectionTimeout(Duration.ofSeconds(5))
                .reconnectWait(Duration.ofSeconds(2))
                .maxReconnects(10)
                .build();
        
        connection = Nats.connect(options);
        
        // Subscribe ke JetStream stream
        // Subject format dari Debezium: product.public.products
        dispatcher = connection.createDispatcher(this::handleMessage);
        dispatcher.subscribe("product.public.products", "product-search-group");
        
        log.info("NATS consumer started, listening on: product.public.products");
    }
    
    private void handleMessage(io.nats.client.Message msg) {
        try {
            String data = new String(msg.getData(), StandardCharsets.UTF_8);
            log.debug("Received CDC event: {}", data);
            
            // Parse Debezium JSON envelope
            JsonNode root = objectMapper.readTree(data);
            JsonNode payload = root.has("payload") ? root.get("payload") : root;
            
            // Debezium envelope structure:
            // { "op": "c|u|d", "before": {...}, "after": {...} }
            String operation = payload.has("op") ? payload.get("op").asText() : "u";
            JsonNode after = payload.has("after") ? payload.get("after") : payload;
            
            switch (operation) {
                case "c", "r" -> { // Create, Read (snapshot)
                    String productId = after.has("id") ? after.get("id").asText() : null;
                    if (productId != null) {
                        log.info("Indexing new product: {}", productId);
                        productIndexService.indexProduct(productId);
                    }
                }
                case "u" -> { // Update
                    String productId = after.has("id") ? after.get("id").asText() : null;
                    if (productId != null) {
                        log.info("Updating product index: {}", productId);
                        productIndexService.indexProduct(productId);
                    }
                }
                case "d" -> { // Delete
                    JsonNode before = payload.has("before") ? payload.get("before") : null;
                    if (before != null && before.has("id")) {
                        String productId = before.get("id").asText();
                        log.info("Removing product from index: {}", productId);
                        productIndexService.removeProduct(productId);
                    }
                }
                default -> log.warn("Unknown operation: {}", operation);
            }
            
            // Ack message
            msg.ack();
            
        } catch (Exception e) {
            log.error("Error processing CDC event", e);
            // Nak untuk retry
            msg.nak();
        }
    }
    
    @PreDestroy
    public void stop() {
        try {
            if (dispatcher != null) {
                connection.closeDispatcher(dispatcher);
            }
            if (connection != null) {
                connection.close();
            }
            log.info("NATS consumer stopped");
        } catch (Exception e) {
            log.error("Error stopping NATS consumer", e);
        }
    }
}
