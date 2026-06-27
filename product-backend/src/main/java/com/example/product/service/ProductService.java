package com.example.product.service;

import com.example.product.dto.ProductDTOs.*;
import com.example.product.entity.Brand;
import com.example.product.entity.Category;
import com.example.product.entity.Product;
import com.example.product.repository.BrandRepository;
import com.example.product.repository.CategoryRepository;
import com.example.product.repository.ProductRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.util.List;
import java.util.UUID;

@Service
@RequiredArgsConstructor
public class ProductService {
    
    private final ProductRepository productRepository;
    private final CategoryRepository categoryRepository;
    private final BrandRepository brandRepository;
    
    @Transactional
    public ProductResponse create(ProductRequest request) {
        // Validate SKU unique
        if (productRepository.existsBySku(request.getSku())) {
            throw new IllegalArgumentException("SKU already exists: " + request.getSku());
        }
        
        // Validate price
        if (request.getPrice().compareTo(BigDecimal.ZERO) < 0) {
            throw new IllegalArgumentException("Price must be greater than or equal to 0");
        }
        
        // Validate stock
        if (request.getStock() != null && request.getStock() < 0) {
            throw new IllegalArgumentException("Stock must be greater than or equal to 0");
        }
        
        // Validate category exists
        Category category = categoryRepository.findById(request.getCategoryId())
                .orElseThrow(() -> new IllegalArgumentException("Category not found: " + request.getCategoryId()));
        
        // Validate brand exists
        Brand brand = brandRepository.findById(request.getBrandId())
                .orElseThrow(() -> new IllegalArgumentException("Brand not found: " + request.getBrandId()));
        
        Product product = Product.builder()
                .sku(request.getSku())
                .name(request.getName())
                .description(request.getDescription())
                .price(request.getPrice())
                .stock(request.getStock() != null ? request.getStock() : 0)
                .imageUrl(request.getImageUrl())
                .category(category)
                .brand(brand)
                .build();
        
        product = productRepository.save(product);
        return toResponse(product);
    }
    
    public Page<ProductResponse> search(String keyword, List<UUID> categoryIds, List<UUID> brandIds,
                                        BigDecimal minPrice, BigDecimal maxPrice, Boolean inStock,
                                        Pageable pageable) {
        return productRepository.searchProducts(keyword, categoryIds, brandIds, 
                minPrice, maxPrice, inStock, pageable)
                .map(this::toResponse);
    }
    
    public ProductResponse findById(UUID id) {
        return productRepository.findById(id)
                .map(this::toResponse)
                .orElseThrow(() -> new IllegalArgumentException("Product not found: " + id));
    }
    
    @Transactional
    public ProductResponse update(UUID id, ProductRequest request) {
        Product product = productRepository.findById(id)
                .orElseThrow(() -> new IllegalArgumentException("Product not found: " + id));
        
        // Validate SKU unique (if changed)
        if (!product.getSku().equals(request.getSku()) && 
            productRepository.existsBySku(request.getSku())) {
            throw new IllegalArgumentException("SKU already exists: " + request.getSku());
        }
        
        // Validate price
        if (request.getPrice().compareTo(BigDecimal.ZERO) < 0) {
            throw new IllegalArgumentException("Price must be greater than or equal to 0");
        }
        
        // Validate category exists
        Category category = categoryRepository.findById(request.getCategoryId())
                .orElseThrow(() -> new IllegalArgumentException("Category not found: " + request.getCategoryId()));
        
        // Validate brand exists
        Brand brand = brandRepository.findById(request.getBrandId())
                .orElseThrow(() -> new IllegalArgumentException("Brand not found: " + request.getBrandId()));
        
        product.setSku(request.getSku());
        product.setName(request.getName());
        product.setDescription(request.getDescription());
        product.setPrice(request.getPrice());
        product.setStock(request.getStock());
        product.setImageUrl(request.getImageUrl());
        product.setCategory(category);
        product.setBrand(brand);
        
        return toResponse(productRepository.save(product));
    }
    
    @Transactional
    public ProductResponse updateStock(UUID id, StockUpdateRequest request) {
        Product product = productRepository.findById(id)
                .orElseThrow(() -> new IllegalArgumentException("Product not found: " + id));
        
        int newStock = product.getStock();
        if (request.getType() == StockUpdateType.INCREASE) {
            newStock += request.getQuantity();
        } else {
            newStock -= request.getQuantity();
            if (newStock < 0) {
                throw new IllegalArgumentException("Stock cannot be negative");
            }
        }
        
        product.setStock(newStock);
        product = productRepository.save(product);
        
        return ProductResponse.builder()
                .id(product.getId())
                .sku(product.getSku())
                .stock(product.getStock())
                .updatedAt(product.getUpdatedAt())
                .build();
    }
    
    @Transactional
    public void delete(UUID id) {
        if (!productRepository.existsById(id)) {
            throw new IllegalArgumentException("Product not found: " + id);
        }
        productRepository.deleteById(id);
    }
    
    private ProductResponse toResponse(Product product) {
        return ProductResponse.builder()
                .id(product.getId())
                .sku(product.getSku())
                .name(product.getName())
                .description(product.getDescription())
                .price(product.getPrice())
                .stock(product.getStock())
                .imageUrl(product.getImageUrl())
                .category(CategoryRef.builder()
                        .id(product.getCategory().getId())
                        .name(product.getCategory().getName())
                        .build())
                .brand(BrandRef.builder()
                        .id(product.getBrand().getId())
                        .name(product.getBrand().getName())
                        .build())
                .createdAt(product.getCreatedAt())
                .updatedAt(product.getUpdatedAt())
                .build();
    }
}
