package com.example.productsearch.repository;

import com.example.productsearch.entity.ProductDocument;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.math.BigDecimal;
import java.util.List;
import java.util.UUID;

@Repository
public interface ProductSearchRepository extends JpaRepository<ProductDocument, UUID> {
    
    /**
     * Full-text search menggunakan PostgreSQL tsvector + tsquery.
     * Lebih akurat daripada LIKE biasa.
     */
    @Query(value = """
        SELECT p.* FROM products p
        WHERE (:keyword IS NULL OR 
            to_tsvector('english', coalesce(p.name, '') || ' ' || coalesce(p.description, '')) 
            @@ plainto_tsquery('english', :keyword))
        ORDER BY ts_rank(
            to_tsvector('english', coalesce(p.name, '') || ' ' || coalesce(p.description, '')),
            plainto_tsquery('english', :keyword)
        ) DESC
        """,
        nativeQuery = true)
    Page<ProductDocument> searchByKeyword(@Param("keyword") String keyword, Pageable pageable);
    
    /**
     * Search dengan filter lengkap + facets support.
     */
    @Query(value = """
        SELECT p.* FROM products p
        WHERE (:keyword IS NULL OR 
            to_tsvector('english', coalesce(p.name, '') || ' ' || coalesce(p.description, '')) 
            @@ plainto_tsquery('english', :keyword))
        AND (:categoryIds IS NULL OR p.category_id = ANY(CAST(:categoryIds AS UUID[])))
        AND (:brandIds IS NULL OR p.brand_id = ANY(CAST(:brandIds AS UUID[])))
        AND (:minPrice IS NULL OR p.price >= :minPrice)
        AND (:maxPrice IS NULL OR p.price <= :maxPrice)
        AND (:inStock IS NULL OR (:inStock = true AND p.stock > 0) OR (:inStock = false AND p.stock = 0))
        """,
        nativeQuery = true)
    Page<ProductDocument> searchWithFilters(
            @Param("keyword") String keyword,
            @Param("categoryIds") String categoryIds,
            @Param("brandIds") String brandIds,
            @Param("minPrice") BigDecimal minPrice,
            @Param("maxPrice") BigDecimal maxPrice,
            @Param("inStock") Boolean inStock,
            Pageable pageable);
    
    /**
     * Facet: count produk per kategori.
     */
    @Query(value = """
        SELECT p.category_id, c.name as category_name, COUNT(*) as count
        FROM products p
        JOIN categories c ON p.category_id = c.id
        WHERE (:keyword IS NULL OR 
            to_tsvector('english', coalesce(p.name, '') || ' ' || coalesce(p.description, '')) 
            @@ plainto_tsquery('english', :keyword))
        AND (:categoryIds IS NULL OR p.category_id NOT IN (SELECT unnest(CAST(:categoryIds AS UUID[]))))
        GROUP BY p.category_id, c.name
        ORDER BY count DESC
        """, nativeQuery = true)
    List<Object[]> getCategoryFacets(
            @Param("keyword") String keyword,
            @Param("categoryIds") String categoryIds);
    
    /**
     * Facet: count produk per brand.
     */
    @Query(value = """
        SELECT p.brand_id, b.name as brand_name, COUNT(*) as count
        FROM products p
        JOIN brands b ON p.brand_id = b.id
        WHERE (:keyword IS NULL OR 
            to_tsvector('english', coalesce(p.name, '') || ' ' || coalesce(p.description, '')) 
            @@ plainto_tsquery('english', :keyword))
        AND (:brandIds IS NULL OR p.brand_id NOT IN (SELECT unnest(CAST(:brandIds AS UUID[]))))
        GROUP BY p.brand_id, b.name
        ORDER BY count DESC
        """, nativeQuery = true)
    List<Object[]> getBrandFacets(
            @Param("keyword") String keyword,
            @Param("brandIds") String brandIds);
}
