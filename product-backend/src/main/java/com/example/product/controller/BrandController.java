package com.example.product.controller;

import com.example.product.dto.ProductDTOs.*;
import com.example.product.service.BrandService;
import lombok.RequiredArgsConstructor;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.web.PageableDefault;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.UUID;

@RestController
@RequestMapping("/api/brands")
@RequiredArgsConstructor
public class BrandController {
    
    private final BrandService brandService;
    
    @PostMapping
    public ResponseEntity<ApiResponse<BrandResponse>> create(@RequestBody BrandRequest request) {
        BrandResponse response = brandService.create(request);
        return ResponseEntity.status(HttpStatus.CREATED)
                .body(ApiResponse.of(response));
    }
    
    @GetMapping
    public ResponseEntity<ApiResponse<Page<BrandResponse>>> findAll(
            @PageableDefault(size = 10) Pageable pageable) {
        Page<BrandResponse> response = brandService.findAll(pageable);
        return ResponseEntity.ok(ApiResponse.of(response));
    }
    
    @GetMapping("/{id}")
    public ResponseEntity<ApiResponse<BrandResponse>> findById(@PathVariable UUID id) {
        BrandResponse response = brandService.findById(id);
        return ResponseEntity.ok(ApiResponse.of(response));
    }
    
    @PutMapping("/{id}")
    public ResponseEntity<ApiResponse<BrandResponse>> update(
            @PathVariable UUID id, @RequestBody BrandRequest request) {
        BrandResponse response = brandService.update(id, request);
        return ResponseEntity.ok(ApiResponse.of(response));
    }
    
    @DeleteMapping("/{id}")
    public ResponseEntity<ApiResponse<Void>> delete(@PathVariable UUID id) {
        brandService.delete(id);
        return ResponseEntity.ok(ApiResponse.of(null));
    }
}
