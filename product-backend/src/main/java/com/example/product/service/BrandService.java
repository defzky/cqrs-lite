package com.example.product.service;

import com.example.product.dto.ProductDTOs.*;
import com.example.product.entity.Brand;
import com.example.product.repository.BrandRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.UUID;

@Service
@RequiredArgsConstructor
public class BrandService {
    
    private final BrandRepository brandRepository;
    
    @Transactional
    public BrandResponse create(BrandRequest request) {
        if (brandRepository.existsByName(request.getName())) {
            throw new IllegalArgumentException("Brand name already exists: " + request.getName());
        }
        
        Brand brand = Brand.builder()
                .name(request.getName())
                .description(request.getDescription())
                .build();
        
        brand = brandRepository.save(brand);
        return toResponse(brand);
    }
    
    public Page<BrandResponse> findAll(Pageable pageable) {
        return brandRepository.findAll(pageable)
                .map(this::toResponse);
    }
    
    public BrandResponse findById(UUID id) {
        return brandRepository.findById(id)
                .map(this::toResponse)
                .orElseThrow(() -> new IllegalArgumentException("Brand not found: " + id));
    }
    
    @Transactional
    public BrandResponse update(UUID id, BrandRequest request) {
        Brand brand = brandRepository.findById(id)
                .orElseThrow(() -> new IllegalArgumentException("Brand not found: " + id));
        
        if (!brand.getName().equals(request.getName()) && 
            brandRepository.existsByName(request.getName())) {
            throw new IllegalArgumentException("Brand name already exists: " + request.getName());
        }
        
        brand.setName(request.getName());
        brand.setDescription(request.getDescription());
        
        return toResponse(brandRepository.save(brand));
    }
    
    @Transactional
    public void delete(UUID id) {
        if (!brandRepository.existsById(id)) {
            throw new IllegalArgumentException("Brand not found: " + id);
        }
        
        try {
            brandRepository.deleteById(id);
        } catch (Exception e) {
            throw new IllegalStateException("Cannot delete brand: products exist");
        }
    }
    
    private BrandResponse toResponse(Brand brand) {
        return BrandResponse.builder()
                .id(brand.getId())
                .name(brand.getName())
                .description(brand.getDescription())
                .createdAt(brand.getCreatedAt())
                .updatedAt(brand.getUpdatedAt())
                .build();
    }
}
