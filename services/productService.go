package services

import (
	"context"
	"time"

	"github.com/souvikjs01/go-ecommerce/model"
	"github.com/souvikjs01/go-ecommerce/request"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProductService interface {
	CreateProduct(userid string, productInfo *request.ProductPayload) (*model.Product, error)
	GetLatestProducts() (*[]model.Product, error)
	DeleteProductsDetails(productId *string) (*model.Product, error)
	UpdateProductsDetails(productId *string, update_product *request.UpdateProductPayload) (*model.Product, error)
	GetProductDetailsByID(productId string) (*model.Product, error)
	GetAllProduct() (*[]model.Product, error)
	GetProductsByQuery(query string) (*[]model.Product, error)
}

type ProductServiceStruct struct {
	db *mongo.Client
}

func NewProductService(db *mongo.Client) *ProductServiceStruct {
	return &ProductServiceStruct{
		db: db,
	}
}

// Create a Product
func (p *ProductServiceStruct) CreateProduct(userId string, productInfo *request.ProductPayload) (*model.Product, error) {
	product_ch := make(chan *model.Product, 32)
	err_ch := make(chan error, 32)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	user_obj_id, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		err_ch <- err
	}

	newProduct := model.NewProduct(
		&(*productInfo).Title,
		&(*productInfo).Desc,
		&(*productInfo).Img,
		&(*productInfo).Categories,
		&(*productInfo).Size,
		&(*productInfo).Color,
		&(*productInfo).Price,
		&(*productInfo).InStock,
		&user_obj_id,
	)

	go func() {
		defer close(product_ch)
		defer close(err_ch)

		// save into the Database
		_, err = p.db.Database("go-ecomm").Collection("products").InsertOne(ctx, newProduct)
		if err != nil {
			err_ch <- err
			return
		}

		product_ch <- newProduct
	}()

	for {
		select {
		case product := <-product_ch:
			return product, nil
		case err := <-err_ch:
			return nil, err
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		}
	}
}

// Get latest Products
func (p *ProductServiceStruct) GetLatestProducts() (*[]model.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	productsChan := make(chan *[]model.Product, 32)
	errChan := make(chan error, 32)

	var products []model.Product

	to_get_latest_products := bson.A{
		bson.M{
			"$sort": bson.M{
				"createdAt": -1,
			},
		},
		bson.M{
			"$limit": 4,
		},
	}

	go func() {
		defer close(errChan)
		defer close(productsChan)

		cur, err := p.db.Database("go-ecomm").Collection("products").Aggregate(ctx, to_get_latest_products)
		if err != nil {
			errChan <- err
			return
		}

		for cur.Next(ctx) {
			var prod model.Product
			err := cur.Decode(&prod)
			if err != nil {
				errChan <- err
				return
			}
			products = append(products, prod)
		}
		productsChan <- &products
	}()

	for {
		select {
		case products := <-productsChan:
			return products, nil
		case err := <-errChan:
			return nil, err
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		}
	}

}

// Delete the Product
func (p *ProductServiceStruct) DeleteProductsDetails(productId *string) (*model.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	productChan := make(chan *model.Product, 32)
	errChan := make(chan error, 32)

	prod_objId, err := primitive.ObjectIDFromHex(*productId)
	if err != nil {
		errChan <- err
	}

	to_delete := bson.M{
		"_id": bson.M{
			"$eq": prod_objId,
		},
	}

	// to delete from the Primary DB
	go func() {
		defer close(productChan)
		defer close(errChan)

		var prod model.Product
		err := p.db.Database("go-ecomm").Collection("products").FindOne(
			ctx,
			bson.M{
				"_id": bson.M{
					"$eq": prod_objId,
				},
			},
		).Decode(&prod)

		if err != nil {
			errChan <- err
			return
		}

		_, err = p.db.Database("go-ecomm").Collection("products").DeleteOne(ctx, to_delete)
		if err != nil {
			errChan <- err
			return
		}
		productChan <- &prod
	}()

	for {
		select {
		case product := <-productChan:
			return product, nil
		case err := <-errChan:
			return nil, err
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		}
	}

}

func (p *ProductServiceStruct) UpdateProductsDetails(product_id *string, update_product *request.UpdateProductPayload) (*model.Product, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	prodChan := make(chan model.Product, 32)
	errChan := make(chan error, 32)

	// var product domain.Product
	prod_objId, err := primitive.ObjectIDFromHex(*product_id)
	if err != nil {
		errChan <- err
	}

	// to update the details of Product in mongodb
	go func() {
		var prod model.Product
		err := p.db.Database("go-ecomm").Collection("products").FindOne(ctx, bson.M{
			"_id": bson.M{
				"$eq": prod_objId,
			},
		}).Decode(&prod)

		if err != nil {
			errChan <- err
			return
		}

		if update_product.Title != nil {
			prod.Title = *update_product.Title
		}
		if update_product.Desc != nil {
			prod.Desc = *update_product.Desc
		}
		if update_product.Img != nil {
			prod.Img = *update_product.Img
		}
		if update_product.Price != nil {
			prod.Price = *update_product.Price
		}
		if update_product.Size != nil {
			prod.Size = *update_product.Size
		}
		if update_product.Categories != nil {
			prod.Categories = *update_product.Categories
		}
		if update_product.Color != nil {
			prod.Color = *update_product.Color
		}
		if update_product.InStock != nil {
			prod.InStock = *update_product.InStock
		}

		_, err = p.db.Database("go-ecomm").Collection("products").UpdateOne(ctx,
			bson.M{
				"_id": bson.M{
					"$eq": prod_objId,
				},
			},
			bson.M{
				"$set": prod,
			},
		)

		if err != nil {
			errChan <- err
			return
		}
		prodChan <- prod
	}()

	for {
		select {
		case product := <-prodChan:
			return &product, nil
		case err := <-errChan:
			return nil, err
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		}
	}
}

// product information by product ID
func (p *ProductServiceStruct) GetProductDetailsByID(productId string) (*model.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	prodChan := make(chan *model.Product, 32)
	errChan := make(chan error, 32)

	prod_obj_id, err := primitive.ObjectIDFromHex(productId)
	if err != nil {
		errChan <- err
	}

	To_query_product := bson.M{
		"_id": bson.M{
			"$eq": prod_obj_id,
		},
	}
	var product model.Product

	go func() {
		defer close(errChan)
		defer close(prodChan)

		err := p.db.Database("go-ecomm").Collection("products").FindOne(ctx, To_query_product).Decode(&product)
		if err != nil {
			errChan <- err
			return
		}

		prodChan <- &product
	}()

	for {
		select {
		case product := <-prodChan:
			return product, nil
		case err := <-errChan:
			return nil, err
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		}
	}

}

// public_product_routes.GET("/query_product")
func (p *ProductServiceStruct) GetProductsByQuery(query string) (*[]model.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	productsChan := make(chan []model.Product, 32)
	errChan := make(chan error, 32)

	// product title, description, price, category,color
	query_filter := bson.M{
		"$and": bson.A{
			bson.M{
				"instock": bson.M{
					"$eq": true,
				},
			},
			bson.M{
				"$or": bson.A{
					bson.M{
						"title": bson.M{
							"$regex":   query,
							"$options": "i",
						},
					},
					bson.M{
						"desc": bson.M{
							"$regex":   query,
							"$options": "i",
						},
					},
					bson.M{
						"price": bson.M{
							"$eq": query,
						},
					},
					bson.M{
						"categories": bson.M{
							"$in": []string{query},
						},
					},
					bson.M{
						"color": bson.M{
							"$in": []string{query},
						},
					},
					bson.M{
						"size": bson.M{
							"$in": []string{query},
						},
					},
				},
			},
		},
	}

	var products []model.Product

	go func() {
		defer close(errChan)
		defer close(productsChan)

		cur, err := p.db.Database("go-ecomm").Collection("products").Find(ctx, query_filter)
		if err != nil {
			errChan <- err
			return
		}
		for cur.Next(ctx) {
			var prod model.Product
			err := cur.Decode(&prod)
			if err != nil {
				errChan <- err
				return
			}
			products = append(products, prod)
		}
		productsChan <- products

	}()

	for {
		select {
		case products = <-productsChan:
			return &products, nil
		case err := <-errChan:
			return nil, err
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		}
	}
}

func (p *ProductServiceStruct) GetAllProduct() (*[]model.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	productsChan := make(chan *[]model.Product, 32)
	errChan := make(chan error, 32)

	var products []model.Product

	go func() {
		defer close(productsChan)
		defer close(errChan)

		cur, err := p.db.Database("go-ecomm").Collection("products").Find(ctx, bson.M{})
		if err != nil {
			errChan <- err
			return
		}

		for cur.Next(ctx) {
			var prod model.Product
			err := cur.Decode(&prod)
			if err != nil {
				errChan <- err
				return
			}
			products = append(products, prod)
		}
		productsChan <- &products
	}()

	for {
		select {
		case products := <-productsChan:
			return products, nil
		case err := <-errChan:
			return nil, err
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		}
	}

}
