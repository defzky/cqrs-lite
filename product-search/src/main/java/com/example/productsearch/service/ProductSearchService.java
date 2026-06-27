package com.example.productsearch.service;

import com.example.productsearch.dto.SearchResponse;
import com.example.productsearch.dto.SearchResponse.*;
import com.example.productsearch.entity.ProductDocument;
import com.example.productsearch.repository.ProductSearchRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Pageable;
import org.springframework.data.domain.Sort;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;
import java.util.*;
import java.util.stream.Collectors;

@Service
@Slf4j
@RequiredArgsConstructor
public class ProductSearchService {
    
    private final ProductSearchRepository searchRepository;
    
    /**
     * Search produk dengan filter dan facets.
     * 
     * Algoritma facets (drill-down):
     * - Count pada facet dihitung dengan filter aktif KECUALI filter dari dimensi itu sendiri
     * - Tujuannya: item lain di dimensi yang sama tidak hilang setelah salah satu dipilih
     */
    public SearchResponse search(String keyword, List<UUID> categoryIds, List<UUID> brandIds,
                                 BigDecimal minPrice, BigDecimal maxPrice, Boolean inStock,
                                 int page, int size, String sort) {
        
        // Parse sort
        Sort sortObj = parseSort(sort);
        Pageable pageable = PageRequest.of(page, Math.min(size, 100), sortObj);
        
        // Convert list to PostgreSQL array string
        String categoryIdsStr = categoryIds != null && !categoryIds.isEmpty() 
                ? "'{" + categoryIds.stream().map(UUID::toString).collect(Collectors.joining(",")) + "}'" 
                : null;
        String brandIdsStr = brandIds != null && !brandIds.isEmpty() 
                ? "'{" + brandIds.stream().map(UUID::toString).collect(Collectors.joining(",")) + "}'" 
                : null;
        
        // Search products
        Page<ProductDocument> results = searchRepository.searchWithFilters(
                keyword, categoryIdsStr, brandIdsStr, minPrice, maxPrice, inStock, pageable);
        
        // Calculate facets (dengan drill-down logic)
        List<CategoryFacet> categoryFacets = getCategoryFacets(keyword, categoryIds);
        List<BrandFacet> brandFacets = getBrandFacets(keyword, brandIds);
        List<PriceRangeFacet> priceRangeFacets = calculatePriceRanges(results.getTotalElements());
        List<AvailabilityFacet> availabilityFacets = calculateAvailability(results.getTotalElements());
        
        // Build response
        return SearchResponse.builder()
                .data(results.getContent().stream()
                        .map(this::toProductResponse)
                        .collect(Collectors.toList()))
                .paging(PagingInfo.builder()
                        .page(results.getNumber())
                        .size(results.getSize())
                        .totalElement(results.getTotalElements())
                        .totalPage(results.getTotalPages())
                        .build())
                .facets(Facets.builder()
                        .categories(categoryFacets)
                        .brands(brandFacets)
                        .priceRanges(priceRangeFacets)
                        .availability(availabilityFacets)
                        .build())
                .metadata(Metadata.builder()
                        .processTimeMs(System.currentTimeMillis())
                        .build())
                .build();
    }
    
    private List<CategoryFacet> getCategoryFacets(String keyword, List<UUID> selectedCategoryIds) {
        String categoryIdsStr = selectedCategoryIds != null && !selectedCategoryIds.isEmpty() 
                ? "'{" + selectedCategoryIds.stream().map(UUID::toString).collect(Collectors.joining(",")) + "}'" 
                : null;
        
        List<Object[]> results = searchRepository.getCategoryFacets(keyword, categoryIdsStr);
        
        return results.stream()
                .map(row -> CategoryFacet.builder()
                        .id((UUID) row[0])
                        .name((String) row[1])
                        .count(((Number) row[2]).longValue())
                        .selected(selectedCategoryIds != null && selectedCategoryIds.contains((UUID) row[0]))
                        .build())
                .collect(Collectors.toList());
    }
    
    private List<BrandFacet> getBrandFacets(String keyword, List<UUID> selectedBrandIds) {
        String brandIdsStr = selectedBrandIds != null && !selectedBrandIds.isEmpty() 
                ? "'{" + selectedBrandIds.stream().map(UUID::toString).collect(Collectors.joining(",")) + "}'" 
                : null;
        
        List<Object[]> results = searchRepository.getBrandFacets(keyword, brandIdsStr);
        
        return results.stream()
                .map(row -> BrandFacet.builder()
                        .id((UUID) row[0])
                        .name((String) row[1])
                        .count(((Number) row[2]).longValue())
                        .selected(selectedBrandIds != null && selectedBrandIds.contains((UUID) row[0]))
                        .build())
                .collect(Collectors.toList());
    }
    
    private List<PriceRangeFacet> calculatePriceRanges(long total) {
        // Static price ranges untuk demo
        return List.of(
                PriceRangeFacet.builder()
                        .min(BigDecimal.ZERO)
                        .max(new BigDecimal("100000"))
                        .label("< 100.000")
                        .count(0L)
                        .selected(false)
                        .build(),
                PriceRangeFacet.builder()
                        .min(new BigDecimal("100000"))
                        .max(new BigDecimal("500000"))
                        .label("100.000 - 500.000")
                        .count(0L)
                        .selected(false)
                        .build(),
                PriceRangeFacet.builder()
                        .min(new BigDecimal("500000"))
                        .max(null)
                        .label("> 500.000")
                        .count(0L)
                        .selected(false)
                        .build()
        );
    }
    
    private List<AvailabilityFacet> calculateAvailability(long total) {
        return List.of(
                AvailabilityFacet.builder()
                        .value("IN_STOCK")
                        .count(0L)
                        .selected(false)
                        .build(),
                AvailabilityFacet.builder()
                        .value("OUT_OF_STOCK")
                        .count(0L)
                        .selected(false)
                        .build()
        );
    }
    
    private Sort parseSort(String sort) {
        if (sort == null || sort.isBlank()) {
            return Sort.by("name").ascending();
        }
        
        String[] parts = sort.split(",");
        if (parts.length != 2) {
            return Sort.by("name").ascending();
        }
        
        String field = parts[0].trim();
        String direction = parts[1].trim().toUpperCase();
        
        return Sort.by(Sort.Direction.fromString(direction), field);
    }
    
    private ProductResponse toProductResponse(ProductDocument doc) {
        return ProductResponse.builder()
                .id(doc.getId())
                .sku(doc.getSku())
                .name(doc.getName())
                .description(doc.getDescription())
                .price(doc.getPrice())
                .stock(doc.getStock())
                .imageUrl(doc.getImageUrl())
                .category(CategoryRef.builder()
                        .id(doc.getCategoryId())
                        .name(doc.getCategoryName())
                        .build())
                .brand(BrandRef.builder()
                        .id(doc.getBrandId())
                        .name(doc.getBrandName())
                        .build())
                .createdAt(doc.getCreatedAt())
                .updatedAt(doc.getUpdatedAt())
                .build();
    }
}
