package com.example.product.controller;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.springframework.data.domain.Page;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class ApiResponse<T> {
    private T data;
    private PagingInfo paging;
    private Metadata metadata;
    
    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    @Builder
    public static class PagingInfo {
        private int page;
        private int size;
        private long totalElement;
        private int totalPage;
    }
    
    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    @Builder
    public static class Metadata {
        private long processTimeMs;
    }
    
    public static <T> ApiResponse<T> of(T data) {
        return ApiResponse.<T>builder()
                .data(data)
                .metadata(Metadata.builder()
                        .processTimeMs(System.currentTimeMillis())
                        .build())
                .build();
    }
    
    public static <T> ApiResponse<T> of(Page<T> page) {
        return ApiResponse.<T>builder()
                .data((T) page.getContent())
                .paging(PagingInfo.builder()
                        .page(page.getNumber())
                        .size(page.getSize())
                        .totalElement(page.getTotalElements())
                        .totalPage(page.getTotalPages())
                        .build())
                .metadata(Metadata.builder()
                        .processTimeMs(System.currentTimeMillis())
                        .build())
                .build();
    }
}
