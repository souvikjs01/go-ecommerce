package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/souvikjs01/go-ecommerce/model"
	"github.com/souvikjs01/go-ecommerce/request"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CartService interface {
	AddToCart(cart *request.AddToCartPayload, userID string) (*model.Cart, error)
	GetCartDetails(userId string) (*model.Cart, error) // debug
	DeleteCart(userId, cartId string) (*model.Cart, error)
	GetAllCarts() (*[]model.Cart, error)
	UpdateCart(cart *model.Cart, userId, cartId string) (*model.Cart, error) // debug
}

type CartServiceStruct struct {
	db *mongo.Client
}

func NewCartService(db *mongo.Client) *CartServiceStruct {
	return &CartServiceStruct{
		db: db,
	}
}

func (c *CartServiceStruct) AddToCart(cart *request.AddToCartPayload, userID string) (*model.Cart, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	cartChan := make(chan *model.Cart, 32)
	errChan := make(chan error, 32)

	user_objId, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		errChan <- err
	}

	productObjID, err := primitive.ObjectIDFromHex(cart.ProductID)
	if err != nil {
		errChan <- err
	}

	// MongoDB
	go func() {
		defer close(errChan)
		defer close(cartChan)

		collection := c.db.Database("go-ecomm").Collection("carts")
		var existingCart model.Cart

		// Check if user already has a cart
		err := collection.FindOne(ctx, bson.M{"userid": user_objId}).Decode(&existingCart)
		if err == mongo.ErrNoDocuments {
			// No cart exists, create a new one
			newCart := &model.Cart{
				ID:     primitive.NewObjectID(),
				UserId: user_objId,
				Products: []model.ProductDetails{
					{
						ProductID: productObjID,
						Quantity:  cart.Quantity,
					},
				},
			}

			_, err = collection.InsertOne(ctx, newCart)
			if err != nil {
				errChan <- fmt.Errorf("failed to create cart: %w", err)
				return
			}
			cartChan <- newCart

		} else if err != nil {
			errChan <- fmt.Errorf("failed to query cart: %w", err)
			return
		}

		// Cart exists, check if product already in cart
		productExists := false
		for _, product := range existingCart.Products {
			if product.ProductID == productObjID {
				productExists = true
				break
			}
		}

		if productExists {
			// Product exists, update quantity
			_, err = collection.UpdateOne(ctx,
				bson.M{
					"_id":                existingCart.ID,
					"userid":             user_objId,
					"products.productid": productObjID,
				},
				bson.M{
					"$inc": bson.M{"products.$.quantity": cart.Quantity},
				},
			)
			if err != nil {
				errChan <- fmt.Errorf("failed to update product quantity: %w", err)
				return
			}
		} else {
			// Product doesn't exist, add to cart
			_, err = collection.UpdateOne(ctx,
				bson.M{
					"_id":    existingCart.ID,
					"userid": user_objId,
				},
				bson.M{
					"$push": bson.M{
						"products": model.ProductDetails{
							ProductID: productObjID,
							Quantity:  cart.Quantity,
						},
					},
				},
			)
			if err != nil {
				errChan <- fmt.Errorf("failed to add product to cart: %w", err)
				return
			}
		}

		// Fetch and return updated cart
		var updatedCart model.Cart
		err = collection.FindOne(ctx, bson.M{"userid": user_objId}).Decode(&updatedCart)
		if err != nil {
			errChan <- fmt.Errorf("failed to fetch updated cart: %w", err)
			return
		}

		cartChan <- &updatedCart
	}()

	for {
		select {
		case cart := <-cartChan:
			return cart, nil
		case err := <-errChan:
			return nil, err
		case <-ctx.Done():
			return nil, fmt.Errorf("context deadline exceeded")
		}
	}
}

// get Cart
func (c *CartServiceStruct) GetCartDetails(userId string) (*model.Cart, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	cartChan := make(chan *model.Cart, 32)
	errChan := make(chan error, 32)

	usrObjID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		errChan <- fmt.Errorf("invalid userId: %w", err)
	}

	filter := bson.M{"userid": usrObjID}

	go func() {
		var cartData model.Cart
		err = c.db.Database("go-ecomm").Collection("carts").FindOne(ctx, filter).Decode(&cartData)
		if err == mongo.ErrNoDocuments {
			errChan <- err
			return
		} else if err != nil {
			errChan <- err
			return
		}

		cartChan <- &cartData
	}()
	for {
		select {
		case cart := <-cartChan:
			return cart, nil
		case err := <-errChan:
			return nil, err
		case <-ctx.Done():
			return nil, fmt.Errorf("context deadline exceeded")
		}
	}
}

// delete Cart
func (c *CartServiceStruct) DeleteCart(userId, cartId string) (*model.Cart, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	cartChan := make(chan *model.Cart, 32)
	errChan := make(chan error, 32)

	var cart model.Cart

	usr_obj_Id, _ := primitive.ObjectIDFromHex(userId)
	cart_obj_Id, _ := primitive.ObjectIDFromHex(cartId)

	query := bson.M{
		"$and": bson.A{
			bson.M{
				"cart_id": cart_obj_Id,
			},
			bson.M{
				"userid": usr_obj_Id,
			},
		},
	}
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := c.db.Database("go-ecomm").Collection("carts").FindOneAndDelete(ctx, query).Decode(&cart)
		if err != nil {
			errChan <- err
			return
		}
		cartChan <- &cart
	}()

	wg.Wait()

	for {
		select {
		case cart := <-cartChan:
			return cart, nil
		case err := <-errChan:
			return nil, err
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		}
	}
}

// Get All carts
func (c *CartServiceStruct) GetAllCarts() (*[]model.Cart, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	query_to_get_all_carts := bson.A{
		bson.M{
			"$sort": bson.M{
				"createdAt": -1,
			},
		},
		bson.M{
			"$limit": 3,
		},
	}

	cartsChan := make(chan *[]model.Cart, 32)
	errChan := make(chan error, 32)
	var carts []model.Cart

	go func() {
		cur, err := c.db.Database("go-ecomm").Collection("carts").Aggregate(ctx, query_to_get_all_carts)
		if err != nil {
			errChan <- err
			return
		}
		for cur.Next(ctx) {
			var cart model.Cart
			err := cur.Decode(&cart)
			if err != nil {
				errChan <- err
				return
			}
			carts = append(carts, cart)
		}

		cartsChan <- &carts
	}()

	for {
		select {
		case carts := <-cartsChan:
			return carts, nil
		case err := <-errChan:
			return nil, err
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		}
	}

}

func (c *CartServiceStruct) UpdateCart(cart *model.Cart, userId, cartId string) (*model.Cart, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cartChan := make(chan *model.Cart, 32)
	errChan := make(chan error, 32)

	usrObjID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		errChan <- fmt.Errorf("invalid userId: %w", err)
	}

	cartObjID, err := primitive.ObjectIDFromHex(cartId)
	if err != nil {
		errChan <- fmt.Errorf("invalid cartId: %w", err)
	}

	filter := bson.M{"_id": cartObjID, "userId": usrObjID}

	go func() {
		var existingCart model.Cart

		// Find the cart
		err := c.db.Database("go-ecomm").Collection("carts").
			FindOne(ctx, filter).Decode(&existingCart)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				errChan <- fmt.Errorf("cart not found for this user")
			} else {
				errChan <- fmt.Errorf("failed to fetch cart: %w", err)
			}
			return
		}

		// Update fields
		existingCart.Products = cart.Products
		existingCart.UserId = usrObjID

		update := bson.M{
			"$set": bson.M{
				"products": existingCart.Products,
			},
		}

		_, err = c.db.Database("go-ecomm").Collection("carts").
			UpdateOne(ctx, filter, update)

		if err != nil {
			errChan <- fmt.Errorf("failed to update cart: %w", err)
			return
		}

		cartChan <- &existingCart
	}()

	select {
	case updated := <-cartChan:
		return updated, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, context.DeadlineExceeded
	}
}
