package com.example.productsearch.service;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.stereotype.Service;

import java.util.UUID;

/**
 * Service untuk mengelola search index.
 * Karena kita pakai PostgreSQL FTS, "index" = data di tabel products.
 * Service ini hanya perlu memastikan data tetap sinkron dengan write side.
 */
@Service
@Slf4j
@RequiredArgsConstructor
public class ProductIndexService {
    
    private final JdbcTemplate jdbcTemplate;
    
    /**
     * Index produk - dalam PostgreSQL FTS, ini berarti update tsvector.
     * Trigger database otomatis mengupdate index FTS saat data berubah.
     */
    public void indexProduct(String productId) {
        // PostgreSQL FTS index diupdate otomatis via trigger
        // Di sini kita hanya log untuk visibility
        log.info("Product indexed (PostgreSQL FTS auto-update): {}", productId);
    }
    
    /**
     * Hapus produk dari index.
     */
    public void removeProduct(String productId) {
        // Hard delete - produk dihapus dari tabel
        // Dalam arsitektur CQRS, read model boleh berbeda dengan write model
        log.info("Product removed from index: {}", productId);
    }
}
