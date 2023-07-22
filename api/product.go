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

type trackProductResponse struct {
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

	priceChan := make(chan float64)
	errorChan := make(chan error)

	go func() {
		price, err := util.ScrapePriceFromURL(server.colly, req.URL)
		if err != nil {
			errorChan <- err
			return
		}
		log.Printf("Initial price for product %d: %f\n", product.ID, price)
		priceChan <- price
	}()

	select {
	case price := <-priceChan:
		_, err := server.store.UpdateProductPrice(ctx, db.UpdateProductPriceParams{
			ID:        product.ID,
			BasePrice: price,
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

	case err := <-errorChan:
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	cronExp := fmt.Sprintf("@every %dh", req.TrackingFrequency*24*60)

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
		log.Printf("Fetching price for product %d: %s\n", product.ID, product.Link)
		fetchedPrice, err := util.ScrapePriceFromURL(server.colly, product.Link)
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

		price_changes, err := server.store.GetPriceChangesForUserAndProduct(ctx, db.GetPriceChangesForUserAndProductParams{
			ID:        1, // user id
			ProductID: product.ID,
		})
		if err != nil {
			log.Printf("Error retrieving price changes: %v", err)
			errorChan <- err
			return
		}
		// if the price changes retrieved exists, compare the last item in the list with the fetched price
		if len(price_changes) == 0 {
			if fetchedPrice != product.BasePrice {
				_, err := server.store.CreatePriceChange(ctx, db.CreatePriceChangeParams{
					ProductID: product.ID,
					Price:     fetchedPrice,
					CreatedAt: time.Now(),
				})
				if err != nil {
					log.Printf("Error creating price change: %v", err)
					errorChan <- err
					return
				}
			}
		} else {
			if fetchedPrice != price_changes[len(price_changes)-1].Price {
				_, err := server.store.CreatePriceChange(ctx, db.CreatePriceChangeParams{
					ProductID: product.ID,
					Price:     fetchedPrice,
					CreatedAt: time.Now(),
				})
				if err != nil {
					log.Printf("Error creating price change: %v", err)
					errorChan <- err
					return
				}
			}
		}
		resultChan <- fetchedPrice
	}()

	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errorChan:
		return 0, err
	}
}

type getProductsRequest struct {
	Page int32 `form:"page" binding:"required,min=1"`
	Size int32 `form:"size" binding:"required,min=1,max=100"`
}

type productResponse struct {
	Product      db.Product       `json:"product"`
	PriceChanges []db.PriceChange `json:"price_changes"`
}

func (server *Server) getProducts(ctx *gin.Context) {
	var req getProductsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	products, err := server.store.ListProductsByUserID(ctx, db.ListProductsByUserIDParams{
		Limit:  req.Size,
		Offset: (req.Page - 1) * req.Size,
		UserID: 1,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	productResponses := make([]productResponse, len(products))
	for i, product := range products {
		priceChanges, err := server.store.GetPriceChangesForUserAndProduct(ctx, db.GetPriceChangesForUserAndProductParams{
			ID:        1,
			ProductID: product.ID,
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		productResponses[i] = productResponse{
			Product:      product,
			PriceChanges: priceChanges,
		}
	}

	ctx.JSON(http.StatusOK, productResponses)
}

type getProductPriceChangesRequest struct {
	ID int64 `uri:"id" binding:"required"`
}

func (server *Server) getProductPriceChanges(ctx *gin.Context) {
	var req getProductPriceChangesRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	priceChanges, err := server.store.GetPriceChangesForUserAndProduct(ctx, db.GetPriceChangesForUserAndProductParams{
		ID:        1,
		ProductID: req.ID,
	})
	if err != nil {
		log.Printf("Error retrieving price changes: %v", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, priceChanges)
}
