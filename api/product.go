package api

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/r-scheele/escout/db/sqlc"
	"github.com/r-scheele/escout/util"
)

type trackProductRequest struct {
	Name                  string  `json:"product_name" binding:"required"`
	URL                   string  `json:"product_url" binding:"required"`
	TrackingFrequency     int32   `json:"tracking_frequency" binding:"required"`
	PercentageChange      float64 `json:"percentage_change" binding:"required"`
	NotificationThreshold float64 `json:"notification_threshold" binding:"required"`
}

func (server *Server) trackProduct(ctx *gin.Context) {
	var req trackProductRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if req.TrackingFrequency <= 0 {
		ctx.JSON(http.StatusBadRequest, errorResponse(fmt.Errorf("tracking frequency must be greater than 0")))
		return
	}

	var product db.Product
	_, err := server.store.GetProductByLinkAndUserID(ctx, db.GetProductByLinkAndUserIDParams{
		UserID: 1,
		Link:   req.URL,
	})
	if err != nil {
		if err != sql.ErrNoRows {
			product, err = server.store.CreateProduct(ctx, db.CreateProductParams{
				UserID:                1,
				Name:                  req.Name,
				Link:                  req.URL,
				BasePrice:             0,
				PercentageChange:      req.PercentageChange,
				TrackingFrequency:     req.TrackingFrequency,
				NotificationThreshold: req.NotificationThreshold,
				CreatedAt:             time.Now(),
			})
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, errorResponse(err))
				return
			}
		}
	} else {
		ctx.JSON(http.StatusBadRequest, errorResponse(fmt.Errorf("product already being tracked")))
		return
	}

	price, err := server.ScrapeProductPrice(ctx, &product)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	log.Printf("Initial price for product %d: %f\n", product.ID, price)

	cronExp := fmt.Sprintf("@every %dm", req.TrackingFrequency)
	err = server.cron.AddFunc(cronExp, func() {
		_, err := server.ScrapeProductPrice(ctx, &product)
		if err != nil {
			log.Printf("Error scraping product price: %v", err)
		}
		// Add additional logic to check price against NotificationThreshold and generate notification
	})
	if err != nil {
		log.Println("Error adding cron job:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to schedule cron job"})
		return
	}

	server.cron.Start()
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Product tracked successfully",
		"product": product,
	})
}

func (server *Server) ScrapeProductPrice(ctx context.Context, product *db.Product) (float64, error) {
	resultChan := make(chan float64)
	errorChan := make(chan error)

	go func() {
		fetchedPrice, err := util.ScrapePriceFromURL(product.Link)
		if err != nil {
			log.Printf("Error fetching price: %v", err)
			errorChan <- err
			return
		}

		if fetchedPrice < 0 {
			log.Printf("Invalid price fetched: %f", fetchedPrice)
			errorChan <- fmt.Errorf("invalid price fetched for product %d: %f", product.ID, fetchedPrice)
			return
		}

		_, err = server.store.UpdateProduct(ctx, db.UpdateProductParams{
			ID:        product.ID,
			BasePrice: fetchedPrice,
		})
		if err != nil {
			log.Printf("Error updating product price: %v", err)
			errorChan <- err
			return
		}

		log.Printf("Updated price for product %d: %f\n", product.ID, fetchedPrice)
		resultChan <- fetchedPrice
	}()

	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errorChan:
		return 0, err
	}
}
