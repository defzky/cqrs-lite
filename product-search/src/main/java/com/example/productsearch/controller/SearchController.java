package com.example.productsearch.controller;

import com.example.productsearch.dto.SearchResponse;
import com.example.productsearch.service.ProductSearchService;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

import java.math.BigDecimal;
import java.util.List;
import java.util.UUID;

@RestController
@RequestMapping("/api/search")
@RequiredArgsConstructor
public class SearchController {
    
    private final ProductSearchService searchService;
    
    @GetMapping("/products")
    public SearchResponse.Response search(
            @RequestParam(required = false) String keyword,
            @RequestParam(required = false) List<UUID> categoryId,
            @RequestParam(required = false) List<UUID> brandId,
            @RequestParam(required = false) BigDecimal minPrice,
            @RequestParam(required = false) BigDecimal maxPrice,
            @RequestParam(required = false) Boolean inStock,
            @RequestParam(defaultValue = "0") int page,
            @RequestParam(defaultValue = "10") int size,
            @RequestParam(defaultValue = "name,asc") String sort) {
        
        return searchService.search(keyword, categoryId, brandId, 
                minPrice, maxPrice, inStock, page, size, sort);
    }
}
