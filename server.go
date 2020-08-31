package main

import (
	"go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"fmt"
	"net/http"
	"time"
	"context"
	"log"
	"encoding/json"
	"strconv"
	// "sync"
)

//Global Variables
// var (
// 	mu sync.Mutex
// )

func helloWorldPingMongodbHandler(w http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
    defer func() {
		if err=client.Disconnect(ctx); err!=nil {
			panic(err)
		}
	}()

	if err=client.Ping(ctx, readpref.Primary()); err!=nil {
		panic(err)
	}

	fmt.Fprintf(w, "Hello World, %s!", request.URL.Path[1:])
}

type Booking struct {
	State string `bson:"state" json:"state"`
	District string `bson:"district" json:"district"`
	Kk string `bson:"kk" json:"kk"` 
	NumMO int `bson:"numMO" json:"numMO"`
	NumMA int `bson:"numMA" json:"numMA"` 
	AvgConsultTime int `bson:"avgConsultTime" json:"avgConsultTime"` 
	MonthlyOpSchedules []DailyOpSchedule `bson:"dailyOpSchedule" json:"dailyOpSchedule"` 
}

type DailyOpSchedule struct {
	Date string `bson:"date" json:"date"`
	StartOpHour int `bson:"startOpHour" json:"startOpHour"`
	EndOpHour int `bson:"endOpHour" json:"endOpHour"` 
	NumMO int `bson:"numMO" json:"numMO"`
	NumMA int `bson:"numMA" json:"numMA"` 
	Queue []int `bson:"queue" json:"queue"` 
	QueueCaps []int `bson:"queueConsultCap" json:"queueCaps"` 
	QueueUsages []int `bson:"queueConsultUsages" json:"queueUsages"` 
}

type DailyOpSchedule2 struct {
	Date string `bson:"date" json:"date"`
	IsHalfDay int `bson:"isHalfDay" json:"isHalfDay"`
	StaffPerDay []int `bson:"staffPerDay" json:"staffPerDay"`
	QueuesCapPerDay []int `bson:"queuesCapPerDay" json:"queuesCapPerDay"` 
	QueuesPerDay []QueuePerHr `bson:"queuesPerDay" json:"queuesPerDay"` 
}

type QueuePerHr struct {
	PatientIds []int `bson:"patientIds" json:"patientIds"`
	BookingReasons []int `bson:"bookingReasons" json:"bookingReasons"`
}

type FindAndUpdateResult struct {
	MonthlyOpSchedules []DailyOpSchedule2 `bson:"monthlyOpSchedule"`
}

func getDailyOpSchedule(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
    defer func() {
		if err=client.Disconnect(ctx); err!=nil {
			panic(err)
		}
		}()
		
	mytcaDB := client.Database("mytca")
	bookingColl := mytcaDB.Collection("booking")
	
	cursor, err := bookingColl.Find(
		ctx,
		bson.D{},
	)
	if err!=nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		// var booking bson.M 
		// var booking bson.D 
		var booking Booking 
		if err = cursor.Decode(&booking); err!=nil {
			log.Fatal(err)
		}
		// fmt.Fprintf(w, "Booking: %v", booking["dailyOpSchedule"])
		// fmt.Fprintf(w, "Raw Output: \n%v\n\n", booking)
		// fmt.Fprintf(w, "%v", booking["dailyOpSchedule"].(bson.A)[1])
		
		output, err := json.MarshalIndent(&booking, "", "\t")
		if err != nil {
			fmt.Fprintf(w, "Error: %v", err)
			return
		}
		// fmt.Fprintf(w, "Json Output: \n%s\n\n", output)
		fmt.Fprintf(w, "%s", output)
		
	}
}
	
func submitHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	r.ParseForm()

	//TODO: This line is an ad-hoc solution to remove the 
	//      std output triggered by CORS preflighted request via http OPTIONS.
	//      The correct solution is to add a reverse-proxy server to the
	//      main server so that all requests are channeled to this reverse proxy
	//      in order to overcome CORS restriction.
	if (len(r.PostForm) == 0) { return } 

	fmt.Println("Form received!")
	fmt.Println(r.PostForm)

	numMO, err := strconv.Atoi(r.PostForm["numMO"][0])
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
    defer func() {
		if err=client.Disconnect(ctx); err!=nil {
			panic(err)
		}
		}()
	mytcaDB := client.Database("mytca")
	bookingColl := mytcaDB.Collection("booking")

	res, err := bookingColl.UpdateOne(
		ctx,
		bson.M{"dailyOpSchedule.date": "2020-08-18"},
		bson.D{
			{"$set", bson.D{{"dailyOpSchedule.$.numMO", numMO}}},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Updated %v MongoDB Documents!\n", res.ModifiedCount)
}

func makeBookingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	r.ParseForm()

	//TODO: This line is an ad-hoc solution to remove the 
	//      std output triggered by CORS preflighted request via http OPTIONS.
	//      The correct solution is to add a reverse-proxy server to the
	//      main server so that all requests are channeled to this reverse proxy
	//      in order to overcome CORS restriction.
	if (len(r.PostForm) == 0) { return } 

	fmt.Println("Clinic: ", r.PostForm["clinic"])
	fmt.Println("Date: ", r.PostForm["date"])
	fmt.Println("OpHourIndex: ", r.PostForm["opHrIdx"])

	wcMajority := writeconcern.New(writeconcern.WMajority(), writeconcern.WTimeout(3*time.Second))
	wcMajorityCollectionOpts := options.Collection().SetWriteConcern(wcMajority)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
    defer func() {
		if err=client.Disconnect(ctx); err!=nil {
			panic(err)
		}
		}()
	mytcaDB := client.Database("test")
	bookingColl := mytcaDB.Collection("booking", wcMajorityCollectionOpts)
	
	queuesCapPerHrPath := fmt.Sprintf("monthlyOpSchedule.$.queuesCapPerDay.%s", r.PostForm["opHrIdx"][0])
	queuesCapPerHrProjPath := "monthlyOpSchedule.$"
	patientIdsPath := fmt.Sprintf("monthlyOpSchedule.$.queuesPerDay.%s.patientIds", r.PostForm["opHrIdx"][0])
	bookingReasonsPath := fmt.Sprintf("monthlyOpSchedule.$.queuesPerDay.%s.bookingReasons", r.PostForm["opHrIdx"][0])
	fmt.Println("queuesCapPerHrPath: ", queuesCapPerHrPath)
	fmt.Println("patientIdsPath: ", patientIdsPath)
	fmt.Println("bookingReasonsPath: ", bookingReasonsPath)
	
	// Step 1: Define the callback that specifies the sequence of operations to perform inside the transaction.
	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Important: You must pass sessCtx as the Context parameter to the operations for them to be executed in the
		// transaction.
		queueCapDecrement := -100
		projection := bson.D{
			{queuesCapPerHrProjPath, 1},
		}
		opts := options.FindOneAndUpdate().SetProjection(projection)
		var updatedDocument FindAndUpdateResult
		err = bookingColl.FindOneAndUpdate(
			sessCtx,
			bson.D{
				{"clinic" , r.PostForm["clinic"][0]},
				{"monthlyOpSchedule.date", r.PostForm["date"][0]},
			},
			bson.D{
				{"$inc", bson.D{{queuesCapPerHrPath, queueCapDecrement}}},
				{"$push", bson.D{{patientIdsPath, 123}}},
				{"$push", bson.D{{bookingReasonsPath, 456}}},
			},
			opts,
		).Decode(&updatedDocument)
		if err != nil {
			// ErrNoDocuments means that the filter did not match any documents in the collection
			if err == mongo.ErrNoDocuments {
				sessCtx.AbortTransaction(sessCtx)
				return "ErrNoDocuments, Transaction Rollbacked!", err
			}
			// log.Fatal(err)
			sessCtx.AbortTransaction(sessCtx)
			return "Transaction Rollbacked!", err
		}
		queueCapIdx, err := strconv.Atoi(r.PostForm["opHrIdx"][0])
		if err != nil {
			// log.Fatal(err)
			sessCtx.AbortTransaction(sessCtx)
			return "Transaction Rollbacked!", err
		}
		updatedQueueCap := updatedDocument.MonthlyOpSchedules[0].QueuesCapPerDay[queueCapIdx] + queueCapDecrement
		fmt.Printf("Updated Document %v\n", updatedQueueCap)

		if updatedQueueCap < 0 {
			fmt.Println("Insufficient Queue Capacity, rolling back transaction...")
			sessCtx.AbortTransaction(sessCtx)
			return "Transaction Rollbacked!", err
		}

		return "Transaction Successful!", nil
	}
	
	// Step 2: Start a session and run the callback using WithTransaction.
	session, err := client.StartSession()
	if err != nil {
		panic(err)
	}
	defer session.EndSession(ctx)
	result, err := session.WithTransaction(ctx, callback)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Transaction Returned Result: %v\n", result)

	resultJson := struct{BookingRes string `json:"bookingRes"`}{
		BookingRes: result.(string),
	}
	outputJson, err := json.MarshalIndent(&resultJson, "", "\t")
		if err != nil {
			fmt.Printf("Json Encoding Error: %v", err)
			return
		}
	fmt.Fprintf(w, "%s", outputJson)
	

	// res, err := bookingColl.UpdateOne(
	// 	ctx,
	// 	bson.D{
	// 		{"clinic" , r.PostForm["clinic"][0]},
	// 		{"monthlyOpSchedule.date", r.PostForm["date"][0]},
	// 	},
	// 	bson.D{
	// 		{"$inc", bson.D{{queuesCapPerHrPath, -1}}},
	// 		{"$push", bson.D{{patientIdsPath, 123}}},
	// 		{"$push", bson.D{{bookingReasonsPath, 456}}},
	// 	},
	// )
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("Updated %v MongoDB Documents!\n", res.ModifiedCount)
}

func main() {
	http.HandleFunc("/", helloWorldPingMongodbHandler)
	http.HandleFunc("/booking", getDailyOpSchedule)
	http.HandleFunc("/submit", submitHandler)
	http.HandleFunc("/book", makeBookingHandler)
	http.ListenAndServe(":8080", nil)
}
	