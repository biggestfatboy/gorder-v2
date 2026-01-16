package adapters

import (
	"context"
	_ "github.com/biggestfatboy/gorder-v2/common/config"
	"github.com/biggestfatboy/gorder-v2/order/domain/order"
	"github.com/biggestfatboy/gorder-v2/order/entity"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type OrderRepositoryMongo struct {
	db *mongo.Client
}

func NewOrderRepositoryMongo(db *mongo.Client) *OrderRepositoryMongo {
	return &OrderRepositoryMongo{db: db}
}

var (
	dbName   = viper.GetString("mongo.db-name")
	collName = viper.GetString("mongo.coll-name")
)

func (o *OrderRepositoryMongo) collection() *mongo.Collection {
	return o.db.Database(dbName).Collection(collName)
}

type orderModel struct {
	MongoID     primitive.ObjectID `bson:"_id"`
	ID          string             `bson:"id"`
	CustomerID  string             `bson:"customer_id"`
	Status      string             `bson:"status"`
	PaymentLink string             `bson:"payment_link"`
	Items       []*entity.Item     `bson:"items"`
}

func (o *OrderRepositoryMongo) Create(ctx context.Context, od *order.Order) (created *order.Order, err error) {
	defer o.logWithTag("create", err, created)
	writeM := o.marshalToModel(od)
	res, err := o.collection().InsertOne(ctx, writeM)
	if err != nil {
		return
	}
	created = od
	created.ID = res.InsertedID.(primitive.ObjectID).Hex()
	return
}

func (o *OrderRepositoryMongo) Get(ctx context.Context, id, _ string) (got *order.Order, err error) {
	defer o.logWithTag("get", err, got)
	read := &orderModel{}
	mongoID, _ := primitive.ObjectIDFromHex(id)

	cond := bson.M{
		"_id": mongoID,
	}
	if err = o.collection().FindOne(ctx, cond).Decode(read); err != nil {
		return
	}
	if read == nil {
		return nil, order.NotFoundError{OrderID: id}
	}
	return o.unmarshalOrder(read), nil
}

func (o *OrderRepositoryMongo) Update(ctx context.Context,
	od *order.Order,
	updateFn func(context.Context, *order.Order,
	) (*order.Order, error)) (err error) {
	defer o.logWithTag("update", err, nil)
	if od == nil {
		panic("got nil order")
	}
	//十五
	session, err := o.db.StartSession()
	if err != nil {
		return
	}
	defer session.EndSession(ctx)
	if err = session.StartTransaction(); err != nil {
		return err
	}
	defer func() {
		if err == nil {
			_ = session.CommitTransaction(ctx)
		} else {
			_ = session.AbortTransaction(ctx)
		}
	}()

	// inside transaction:
	oldOrder, err := o.Get(ctx, od.ID, od.CustomerID)
	if err != nil {
		return
	}
	updated, err := updateFn(ctx, od)
	if err != nil {
		return
	}
	mongoID, _ := primitive.ObjectIDFromHex(od.ID)
	res, err := o.collection().UpdateOne(
		ctx,
		bson.M{
			"_id":         mongoID,
			"customer_id": oldOrder.CustomerID,
		},
		bson.M{
			"$set": bson.M{
				"status":       updated.Status,
				"payment_link": updated.PaymentLink,
			},
		},
	)
	if err != nil {
		return
	}
	o.logWithTag("finish_update", err, res)
	return
}

func (o *OrderRepositoryMongo) logWithTag(tag string, err error, result interface{}) {
	l := logrus.WithFields(logrus.Fields{
		"tag":            "order_repository_mongo",
		"performed_time": time.Now().Unix(),
		"err":            err,
		"result":         result,
	})
	if err != nil {
		l.Infof("%s_fail", tag)
	} else {
		l.Infof("%s_success", tag)
	}
}

func (o *OrderRepositoryMongo) marshalToModel(od *order.Order) *orderModel {
	return &orderModel{
		MongoID:     primitive.NewObjectID(),
		ID:          od.ID,
		CustomerID:  od.CustomerID,
		Status:      od.Status,
		PaymentLink: od.PaymentLink,
		Items:       od.Items,
	}
}

func (o *OrderRepositoryMongo) unmarshalOrder(od *orderModel) *order.Order {
	return &order.Order{
		ID:          od.MongoID.Hex(),
		CustomerID:  od.CustomerID,
		Status:      od.Status,
		PaymentLink: od.PaymentLink,
		Items:       od.Items,
	}
}
