package services

import (
	"context"
	"fmt"
	"time"

	"github.com/souvikjs01/go-ecommerce/model"
	"github.com/souvikjs01/go-ecommerce/request"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderService interface {
	CreateOrder(order *request.CreateOrderPayload, userId string) (*model.Order, error)
	GetUserOrders(userID string) (*[]model.Order, error)
	// GetAllOrders() (*[]model.Order, error)
	// DeleteUserOrder(userId, orderId string) (*model.Order, error)
	// UpdateOrderDetails(order *model.Order, userid, orderid string) (*model.Order, error)
}

type OrderServiceStruct struct {
	db *mongo.Client
}

func NewOrderService(db *mongo.Client) *OrderServiceStruct {
	return &OrderServiceStruct{
		db: db,
	}
}

func (o *OrderServiceStruct) CreateOrder(order *request.CreateOrderPayload, userId string) (*model.Order, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	orderChan := make(chan *model.Order, 32)
	errChan := make(chan error, 32)

	userObjID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		errChan <- err
	}

	go func() {
		defer close(errChan)
		defer close(orderChan)

		totalAmount := 0
		for _, product := range order.Products {
			if product.ProductID.IsZero() {
				errChan <- fmt.Errorf("invalid product ID: %v", product.ProductID)
				return
			}

			// Fetch product from DB
			var prod model.Product
			err := o.db.Database("go-ecomm").Collection("products").FindOne(ctx, bson.M{
				"_id": product.ProductID,
			}).Decode(&prod)
			if err != nil {
				errChan <- fmt.Errorf("product %v not found", product.ProductID.Hex())
				return
			}

			totalAmount += (prod.Price * product.Quantity)
			// Check stock
			if !prod.InStock {
				errChan <- fmt.Errorf("product %s is out of stock", prod.Title)
				return
			}
		}

		newOrderStruct := model.Order{
			UserId:   userObjID,
			Amount:   totalAmount,
			Status:   order.Status,
			Address:  order.Address,
			Products: order.Products,
		}

		createNewOrder := model.NewOrder(&newOrderStruct)

		_, err = o.db.Database("go-ecomm").Collection("orders").InsertOne(ctx, createNewOrder)
		if err != nil {
			errChan <- err
			return
		}
		orderChan <- createNewOrder
	}()

	for {
		select {
		case err := <-errChan:
			return nil, err
		case order := <-orderChan:
			return order, nil
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		}
	}
}

func (o *OrderServiceStruct) GetUserOrders(userID string) (*[]model.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	ordersChan := make(chan *[]model.Order)
	errChan := make(chan error, 32)

	userObjId, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid userId")
	}

	go func() {
		cur, err := o.db.Database("go-ecomm").Collection("orders").Find(ctx, bson.M{"userId": userObjId})
		if err != nil {
			errChan <- err
			return
		}
		defer cur.Close(ctx)

		var orders []model.Order
		for cur.Next(ctx) {
			var order model.Order
			if err := cur.Decode(&order); err != nil {
				errChan <- err
			}
			orders = append(orders, order)
		}
		ordersChan <- &orders
	}()

	select {
	case err := <-errChan:
		return nil, err
	case orders := <-ordersChan:
		return orders, nil
	case <-ctx.Done():
		return nil, context.DeadlineExceeded
	}

}

// func (o *OrderServiceStruct) GetAllOrders() (*[]model.Order, error) {
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
// 	defer cancel()

// 	ordersChan := make(chan *[]model.Order)
// 	errChan := make(chan error, 32)

// 	var orders []model.Order
// 	to_get_orders := bson.A{
// 		bson.M{
// 			"$sort": bson.M{
// 				"createdAt": -1,
// 			},
// 		},
// 		bson.M{
// 			"$limit": 3,
// 		},
// 	}
// 	// from MongoDB
// 	go func() {
// 		cur, err := o.db.Database("go-ecomm").Collection("orders").Aggregate(ctx, to_get_orders)
// 		if err != nil {
// 			errChan <- err
// 			return
// 		}
// 		for cur.Next(ctx) {
// 			var order model.Order
// 			err := cur.Decode(&order)
// 			if err != nil {
// 				errChan <- err
// 				return
// 			}
// 			orders = append(orders, order)
// 		}

// 		ordersChan <- &orders
// 	}()

// 	for {
// 		select {
// 		case err := <-errChan:
// 			return nil, err
// 		case orders := <-ordersChan:
// 			return orders, nil
// 		case <-ctx.Done():
// 			return nil, context.DeadlineExceeded
// 		}
// 	}
// }

// func (o *OrderServiceStruct) DeleteUserOrder(userId, orderId string) (*model.Order, error) {
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
// 	defer cancel()

// 	orderChan := make(chan *model.Order, 32)
// 	errChan := make(chan error, 32)

// 	user_obj_id, _ := primitive.ObjectIDFromHex(userId)
// 	order_obj_id, _ := primitive.ObjectIDFromHex(orderId)

// 	var wg sync.WaitGroup

// 	var order model.Order

// 	del_query := bson.M{
// 		"$and": bson.A{
// 			bson.M{"userid": user_obj_id},
// 			bson.M{"_id": order_obj_id},
// 		},
// 	}

// 	wg.Add(1)
// 	// delete order from the mongodb
// 	go func() {
// 		defer func() {
// 			wg.Done()
// 		}()
// 		err := o.db.Database("go-ecomm").Collection("orders").FindOne(ctx, bson.M{
// 			"_id":    order_obj_id,
// 			"userid": user_obj_id,
// 		}).Decode(&order)

// 		if err != nil {
// 			errChan <- err
// 			return
// 		}

// 		_, err = o.db.Database("go-ecomm").Collection("orders").DeleteOne(ctx, del_query)
// 		if err != nil {
// 			errChan <- err
// 			return
// 		}
// 		orderChan <- &order
// 	}()

// 	wg.Wait()

// 	for {
// 		select {
// 		case err := <-errChan:
// 			return nil, err
// 		case order := <-orderChan:
// 			return order, nil
// 		case <-ctx.Done():
// 			return nil, context.DeadlineExceeded
// 		}
// 	}
// }

// func (o *OrderServiceStruct) UpdateOrderDetails(u_order *model.Order, userid, orderid string) (*model.Order, error) {
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
// 	defer cancel()

// 	orderChan := make(chan *model.Order, 32)
// 	errChan := make(chan error, 32)

// 	var updatedOrder model.Order
// 	var wg sync.WaitGroup

// 	usr_Obj_Id, _ := primitive.ObjectIDFromHex(userid)
// 	order_Obj_Id, _ := primitive.ObjectIDFromHex(orderid)

// 	to_find_query := bson.M{
// 		"_id":    order_Obj_Id,
// 		"userid": usr_Obj_Id,
// 	}

// 	wg.Add(1)
// 	// Update the User Order Details in MongoDb
// 	go func() {
// 		defer func() {
// 			wg.Done()
// 		}()
// 		err := o.db.Database("go-ecomm").Collection("orders").FindOne(ctx, to_find_query).Decode(&updatedOrder)
// 		if err != nil {
// 			errChan <- err
// 			return
// 		}

// 		if len((*u_order).Products) != 0 {
// 			updatedOrder.Products = (*u_order).Products
// 		}
// 		if (*u_order).Amount != 0 {
// 			updatedOrder.Amount = (*u_order).Amount
// 		}

// 		if (*u_order).Status != "" {
// 			updatedOrder.Status = (*u_order).Status
// 		}
// 		_, err = o.db.Database("go-ecomm").Collection("orders").UpdateOne(ctx, to_find_query, bson.M{
// 			"$set": updatedOrder,
// 		})
// 		if err != nil {
// 			errChan <- err
// 			return
// 		}
// 		orderChan <- &updatedOrder
// 	}()

// 	wg.Wait()

// 	for {
// 		select {
// 		case err := <-errChan:
// 			return nil, err
// 		case updatedOrderDetails := <-orderChan:
// 			return updatedOrderDetails, nil
// 		case <-ctx.Done():
// 			return nil, context.DeadlineExceeded
// 		}
// 	}
// }
