package com.example.productsearch.dto;

import lombok.*;
import java.math.BigDecimal;
import java.time.Instant;
import java.util.List;
import java.util.UUID;

public class SearchResponse {
    
    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    @Builder
    public static class Response {
        private List<ProductResponse> data;
        private PagingInfo paging;
        private Facets facets;
        private Metadata metadata;
    }
    
    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    @Builder
    public static class ProductResponse {
        private UUID id;
        private String sku;
        private String name;
        private String description;
        private BigDecimal price;
        private Integer stock;
        private String imageUrl;
        private CategoryRef category;
        private BrandRef brand;
        private Instant createdAt;
        private Instant updatedAt;
    }
    
    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    @Builder
    public static class CategoryRef {
        private UUID id;
        private String name;
    }
    
    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    @Builder
    public static class BrandRef {
        private UUID id;
        private String name;
    }
    
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
    public static class Facets {
        private List<CategoryFacet> categories;
        private List<BrandFacet> brands;
        private List<PriceRangeFacet> priceRanges;
        private List<AvailabilityFacet> availability;
    }
    
    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    @Builder
    public static class CategoryFacet {
        private UUID id;
        private String name;
        private Long count;
        private Boolean selected;
    }
    
    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    @Builder
    public static class BrandFacet {
        private UUID id;
        private String name;
        private Long count;
        private Boolean selected;
    }
    
    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    @Builder
    public static class PriceRangeFacet {
        private BigDecimal min;
        private BigDecimal max;
        private String label;
        private Long count;
        private Boolean selected;
    }
    
    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    @Builder
    public static class AvailabilityFacet {
        private String value;
        private Long count;
        private Boolean selected;
    }
    
    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    @Builder
    public static class Metadata {
        private long processTimeMs;
    }
}
