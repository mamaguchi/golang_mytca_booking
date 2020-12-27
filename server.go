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
	"mytca/booking/db/my_mongodb"
	"mytca/booking/db/my_ldap"
	"mytca/booking/auth"
)


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

type Booking2 struct {
	State string `bson:"state" json:"state"`
	District string `bson:"district" json:"district"`
	Kk string `bson:"kk" json:"kk"` 
	NumMO int `bson:"numMO" json:"numMO"`
	NumMA int `bson:"numMA" json:"numMA"` 
	AvgConsultTime int `bson:"avgConsultTime" json:"avgConsultTime"` 
	MonthlyOpSchedules []DailyOpSchedule `bson:"dailyOpSchedule" json:"dailyOpSchedule"` 
}

// type DailyOpSchedule struct {
// 	Date string `bson:"date" json:"date"`
// 	StartOpHour int `bson:"startOpHour" json:"startOpHour"`
// 	EndOpHour int `bson:"endOpHour" json:"endOpHour"` 
// 	NumMO int `bson:"numMO" json:"numMO"`
// 	NumMA int `bson:"numMA" json:"numMA"` 
// 	Queue []int `bson:"queue" json:"queue"` 
// 	QueueCaps []int `bson:"queueConsultCap" json:"queueCaps"` 
// 	QueueUsages []int `bson:"queueConsultUsages" json:"queueUsages"` 
// }

type DailyOpSchedule2 struct {
	Date string `bson:"date" json:"date"`
	IsHalfDay int `bson:"isHalfDay" json:"isHalfDay"`
	StaffPerDay []int `bson:"staffPerDay" json:"staffPerDay"`
	QueuesCapPerDay []int `bson:"queuesCapPerDay" json:"queuesCapPerDay"` 
	QueuesPerDay []QueuePerHr `bson:"queuesPerDay" json:"queuesPerDay"` 
}

type DailyOpSchedule3 struct {
	// Date string 				`bson:"date" json:"date"`
	// StaffPerDay []int 			`bson:"staffPerDay" json:"staffPerDay"`
	DayOfWeek int				`bson:"dayOfWeek" json:"dayOfWeek"`
	QueuesCapPerDay []int 		`bson:"queuesCapPerDay" json:"queuesCapPerDay"` 
	QueuesUsgPerDay []int 		`bson:"queuesUsgPerDay" json:"queuesUsgPerDay"` 
	// QueuesPerDay []QueuePerHr 	`bson:"queuesPerDay" json:"queuesPerDay"` 
}

type ClinicSvcMongodbQueueMeta struct {
	DayOfWeek int				`bson:"dayOfWeek" json:"dayOfWeek"`
	QueuesCapPerDay []int 		`bson:"queuesCapPerDay" json:"queuesCapPerDay"` 
	QueuesUsgPerDay []int 		`bson:"queuesUsgPerDay" json:"queuesUsgPerDay"` 
}

// type QueuePerHr struct {
// 	PatientIds []int `bson:"patientIds" json:"patientIds"`
// 	BookingReasons []int `bson:"bookingReasons" json:"bookingReasons"`
// }

type ClinicSvcMongodbMeta struct {
	MonthlyOpSchedules []DailyOpSchedule3 `bson:"monthlyOpSchedule"`
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
		var booking Booking 
		if err = cursor.Decode(&booking); err!=nil {
			log.Fatal(err)
		}
		
		outputJson, err := json.MarshalIndent(&booking, "", "\t")
		if err != nil {
			fmt.Fprintf(w, "Error: %v", err)
			return
		}
		fmt.Fprintf(w, "%s", outputJson)
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
		var updatedDocument ClinicSvcMongodbMeta
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

func getClinicsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	clinicDirJson, err := my_ldap.GetAllClinics()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "%s", clinicDirJson)
}

func makeBookingHandler2(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	r.ParseForm()

	//TODO: This line is an ad-hoc solution to remove the 
	//      std output triggered by CORS preflighted request via http OPTIONS.
	//      The correct solution is to add a reverse-proxy server to the
	//      main server so that all requests are channeled to this reverse proxy
	//      in order to overcome CORS restriction.
	if (len(r.PostForm) == 0) { return } 

	fmt.Println("ClinicId: ", r.PostForm["clinicId"])
	fmt.Println("Date: ", r.PostForm["date"])
	fmt.Println("OpHourIndex: ", r.PostForm["opHrIdx"])

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()	
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	defer func() {
		if err=client.Disconnect(ctx); err!=nil {
			panic(err)
		}
	}()
	mytcaDB := client.Database("test")
	bookingColl := mytcaDB.Collection("booking2")

	// queuesCapPerHrPath := fmt.Sprintf("monthlyOpSchedule.$.queuesCapPerDay.%s", r.PostForm["opHrIdx"][0])
	// queuesCapPerHrProjPath := "monthlyOpSchedule.$"
	queuesCapPerHrProjPath := "monthlyOpSchedule"
	queuesUsgPerHrPath := fmt.Sprintf("monthlyOpSchedule.$.queuesUsgPerDay.%s", r.PostForm["opHrIdx"][0])
	patientIdsPath := fmt.Sprintf("monthlyOpSchedule.$.queuesPerDay.%s.patientIds", r.PostForm["opHrIdx"][0])
	bookingReasonsPath := fmt.Sprintf("monthlyOpSchedule.$.queuesPerDay.%s.bookingReasons", r.PostForm["opHrIdx"][0])

	// Step 1: Define the callback that specifies the sequence of operations to perform inside the transaction.
	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Important: You must pass sessCtx as the Context parameter to the operations for them to be executed in the
		// transaction.
		queueUsgIncrement := 100
		// projection := bson.D{
		// 	{queuesCapPerHrProjPath, 1},
		// }

		projection := bson.D{
			{queuesCapPerHrProjPath, 
				bson.D{
					{"$elemMatch", bson.D{{"date", r.PostForm["date"][0]}}},
				},
			},		
		}

		opts := options.FindOneAndUpdate()
		opts.SetProjection(projection)
		opts.SetReturnDocument(options.After)
		var updatedDocument ClinicSvcMongodbMeta
		err = bookingColl.FindOneAndUpdate(
			sessCtx,
			bson.D{
				{"clinic" , r.PostForm["clinicId"][0]},
				{"service", r.PostForm["service"][0]},
				{"monthlyOpSchedule.date", r.PostForm["date"][0]},
			},
			bson.D{
				{"$inc", bson.D{{queuesUsgPerHrPath, queueUsgIncrement}}},
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
		updatedQueueCap := updatedDocument.MonthlyOpSchedules[0].QueuesCapPerDay[queueCapIdx]
		updatedQueueUsg := updatedDocument.MonthlyOpSchedules[0].QueuesUsgPerDay[queueCapIdx]
		fmt.Printf("Updated QueuesUsgPerDay to: %v\n", updatedQueueUsg)

		if updatedQueueUsg > updatedQueueCap {
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

	// Write result back to web service client
	resultJson := struct{BookingRes string `json:"bookingRes"`}{
		BookingRes: result.(string),
	}
	outputJson, err := json.MarshalIndent(&resultJson, "", "\t")
	if err != nil {
		fmt.Printf("Json Encoding Error: %v", err)
		return
	}
	fmt.Fprintf(w, "%s", outputJson)


	// queueUsgIncrement := 100
	// // opts := options.Update().SetUpsert(true)
	// res, err := bookingColl.UpdateOne(
	// 	ctx, 
	// 	bson.D{
	// 		// {"clinic" , r.PostForm["clinicId"][0]},
	// 		// {"monthlyOpSchedule.date", r.PostForm["date"][0]},
	// 		// {"monthlyOpSchedule.queuesCapPerDay", 
	// 		// 	bson.D{
	// 		// 		{"$gte", 200},
	// 		// 	},
	// 		// },

	// 		{"$and", bson.A{
	// 			bson.D{{"clinic" , r.PostForm["clinicId"][0]}},
	// 			bson.D{{"monthlyOpSchedule.date", r.PostForm["date"][0]}},
	// 			bson.D{{queuesCapPerHrPath, 
	// 				bson.D{
	// 					{"$gte", 200},
	// 				},
	// 			}},
	// 		}},
	// 	},
	// 	bson.D{
	// 		{"$inc", bson.D{{queuesUsgPerHrPath, queueUsgIncrement}}},
	// 		{"$push", bson.D{{patientIdsPath, 123}}},
	// 		{"$push", bson.D{{bookingReasonsPath, 456}}},
	// 	},		
	// )
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("%d of documents matched by the filter\n", res.MatchedCount)
	// fmt.Printf("%d of documents modified by the operation\n", res.ModifiedCount)
	// fmt.Printf("%d of documents upserted by the operation\n", res.UpsertedCount)
	// fmt.Printf("The _id field of the upserted document, or nil if no upsert was done: %v\n", res.UpsertedID)

	// return
}

func makeBookingHandler3(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	r.ParseForm()

	//TODO: This line is an ad-hoc solution to remove the 
	//      std output triggered by CORS preflighted request via http OPTIONS.
	//      The correct solution is to add a reverse-proxy server to the
	//      main server so that all requests are channeled to this reverse proxy
	//      in order to overcome CORS restriction.
	if (len(r.PostForm) == 0) { return } 

	fmt.Println("ClinicId: ", r.PostForm["clinicId"])
	fmt.Println("Date: ", r.PostForm["date"])
	fmt.Println("OpHourIndex: ", r.PostForm["opHrIdx"])

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()	
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	defer func() {
		if err=client.Disconnect(ctx); err!=nil {
			panic(err)
		}
	}()
	mytcaDB := client.Database("test")
	bookingColl := mytcaDB.Collection("booking3")

	dailyOpScheduleProjPath := "monthlyOpSchedule"
	queuesUsgPerHrPath := fmt.Sprintf("monthlyOpSchedule.$.queuesUsgPerDay.%s", r.PostForm["opHrIdx"][0])
	patientIdsPath := fmt.Sprintf("monthlyOpSchedule.$.queuesPerDay.%s.patientIds", r.PostForm["opHrIdx"][0])
	bookingReasonsPath := fmt.Sprintf("monthlyOpSchedule.$.queuesPerDay.%s.bookingReasons", r.PostForm["opHrIdx"][0])

	// Step 1: Define the callback that specifies the sequence of operations to perform inside the transaction.
	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Important: You must pass sessCtx as the Context parameter to the operations for them to be executed in the
		// transaction.
		queueUsgIncrement := 100		

		projection := bson.D{
			{dailyOpScheduleProjPath, 
				bson.D{
					{"$elemMatch", bson.D{{"date", r.PostForm["date"][0]}}},
				},
			},		
		}

		opts := options.FindOneAndUpdate()
		opts.SetProjection(projection)
		opts.SetReturnDocument(options.After)
		var updatedDocument ClinicSvcMongodbMeta
		err = bookingColl.FindOneAndUpdate(
			sessCtx,
			bson.D{
				{"clinic" , r.PostForm["clinicId"][0]},
				{"dept", "opd_disease"},
				{"monthlyOpSchedule.date", r.PostForm["date"][0]},
			},
			bson.D{
				{"$inc", bson.D{{queuesUsgPerHrPath, queueUsgIncrement}}},
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
		updatedQueueCap := updatedDocument.MonthlyOpSchedules[0].QueuesCapPerDay[queueCapIdx]
		updatedQueueUsg := updatedDocument.MonthlyOpSchedules[0].QueuesUsgPerDay[queueCapIdx]
		fmt.Printf("Updated QueuesUsgPerDay to: %v\n", updatedQueueUsg)

		if updatedQueueUsg > updatedQueueCap {
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

	// Write result back to web service client
	resultJson := struct{BookingRes string `json:"bookingRes"`}{
		BookingRes: result.(string),
	}
	outputJson, err := json.MarshalIndent(&resultJson, "", "\t")
	if err != nil {
		fmt.Printf("Json Encoding Error: %v", err)
		return
	}
	fmt.Fprintf(w, "%s", outputJson)
}

type SvcExistCheck struct {
	SvcsChecked []SvcChecked	`json:"svcsChecked"`
}

type SvcChecked struct {
	Name string		`json:"svcName"`
	IfExist bool	`json:"ifExist"`
}

func checkIfClinicSvcExistHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	r.ParseForm()

	//TODO: This line is an ad-hoc solution to remove the 
	//      std output triggered by CORS preflighted request via http OPTIONS.
	//      The correct solution is to add a reverse-proxy server to the
	//      main server so that all requests are channeled to this reverse proxy
	//      in order to overcome CORS restriction.
	if (len(r.PostForm) == 0) { return } 
	
	// fmt.Println("Post data: \n", r.PostForm)
	clinic := r.PostForm["clinicId"][0]
	district := r.PostForm["district"][0]
	state := r.PostForm["state"][0]
	
	svcsToCheck := []string{
		// SVC_FUNDOSCOPY,
		// SVC_XRAY,
		my_ldap.SVC_FUNDOSCOPY,
		my_ldap.SVC_XRAY,
	}
	var lstOfSvcsChecked SvcExistCheck
	for _, svc := range svcsToCheck {
		dept := my_ldap.SvcToDeptMap[svc]
		ifSvcExist, err := checkIfSvcExist(svc, dept, clinic, district, state)
		if err != nil {
			log.Print(err)
			continue
		}
		svcChecked := SvcChecked{Name: svc, IfExist: ifSvcExist}
		lstOfSvcsChecked.SvcsChecked = append(lstOfSvcsChecked.SvcsChecked, svcChecked)
		// fmt.Printf("If Service '%s' Exists: %t \n", svc, ifSvcExist)			
	}
	// fmt.Printf("List of services checked: %+v \n", lstOfSvcsChecked)
	outputJson, err := json.MarshalIndent(lstOfSvcsChecked, "", "\t")
	if err != nil {
		log.Print(err)
	}
	fmt.Fprintf(w, "%s", outputJson)
}

func getClinicSvcMetaHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	r.ParseForm()

	//TODO: This line is an ad-hoc solution to remove the 
	//      std output triggered by CORS preflighted request via http OPTIONS.
	//      The correct solution is to add a reverse-proxy server to the
	//      main server so that all requests are channeled to this reverse proxy
	//      in order to overcome CORS restriction.
	if (len(r.PostForm) == 0) { return } 

	service := r.PostForm["service"][0]
	clinicId := r.PostForm["clinicId"][0]
	district := r.PostForm["district"][0]
	state := r.PostForm["state"][0]

	clinicServiceMeta, err := GetClinicServiceMeta(service, clinicId, district, state)
	if err != nil {
		log.Print(err)
		w.WriteHeader(500)
		fmt.Fprintf(w, "Internal server error!")
		return
	}

	clinicServiceMetaJson, err := json.MarshalIndent(&clinicServiceMeta, "", "\t")
    if err != nil {
		log.Print(err)
		w.WriteHeader(500)
		fmt.Fprintf(w, "Internal server error!")
		return
	}
	fmt.Fprintf(w, "%s", clinicServiceMetaJson)
}

func getClinicSvcMetaHandler2(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	r.ParseForm()

	//TODO: This line is an ad-hoc solution to remove the 
	//      std output triggered by CORS preflighted request via http OPTIONS.
	//      The correct solution is to add a reverse-proxy server to the
	//      main server so that all requests are channeled to this reverse proxy
	//      in order to overcome CORS restriction.
	if (len(r.PostForm) == 0) { return } 

	//TODO: To create a service-to-department mapping.
	dept := "opd_disease"

	service := r.PostForm["service"][0]
	clinicId := r.PostForm["clinicId"][0]
	district := r.PostForm["district"][0]
	state := r.PostForm["state"][0]

	clinicServiceMeta, err := GetClinicSvcOpHrs(service, dept, clinicId, district, state)
	if err != nil {
		log.Print(err)
		w.WriteHeader(500)
		fmt.Fprintf(w, "Internal server error!")
		return
	}

	clinicServiceMetaJson, err := json.MarshalIndent(&clinicServiceMeta, "", "\t")
    if err != nil {
		log.Print(err)
		w.WriteHeader(500)
		fmt.Fprintf(w, "Internal server error!")
		return
	}
	fmt.Fprintf(w, "%s", clinicServiceMetaJson)
}

func getClinicSvcQueueMetaHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	r.ParseForm()

	//TODO: This line is an ad-hoc solution to remove the 
	//      std output triggered by CORS preflighted request via http OPTIONS.
	//      The correct solution is to add a reverse-proxy server to the
	//      main server so that all requests are channeled to this reverse proxy
	//      in order to overcome CORS restriction.
	if (len(r.PostForm) == 0) { return } 

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()	
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	defer func() {
		if err=client.Disconnect(ctx); err!=nil {
			panic(err)
		}
	}()
	mytcaDB := client.Database("test")
	bookingColl := mytcaDB.Collection("booking2")

	queuesCapPerHrProjPath := "monthlyOpSchedule"
	projection := bson.D{
		{queuesCapPerHrProjPath, 
			bson.D{
				{"$elemMatch", bson.D{{"date", r.PostForm["date"][0]}}},
			},
		},		
	}

	opts := options.FindOne()
	opts.SetProjection(projection)
	var clinicSvcMongodbMeta ClinicSvcMongodbMeta
	// var clinicSvcMongodbQueueMeta ClinicSvcMongodbQueueMeta
	var clinicSvcMongodbQueueMeta DailyOpSchedule3
	err = bookingColl.FindOne(
		ctx,
		bson.D{
			{"clinic" , r.PostForm["clinicId"][0]},
			{"service", r.PostForm["service"][0]},
			{"monthlyOpSchedule.date", r.PostForm["date"][0]},
		},
		opts,
	).Decode(&clinicSvcMongodbMeta)
	if err != nil {
		log.Print(err)
		w.WriteHeader(500)
		fmt.Fprintf(w, "Internal server error!")
		return
	}
	clinicSvcMongodbQueueMeta = clinicSvcMongodbMeta.MonthlyOpSchedules[0]
	// fmt.Println("ClinicSvcMongodbQueueMeta: \n", clinicSvcMongodbQueueMeta)

	outputJson, err := json.MarshalIndent(&clinicSvcMongodbQueueMeta, "", "\t")
	if err != nil {
		log.Print(err)
		w.WriteHeader(500)
		fmt.Fprintf(w, "Internal server error!")
		return
	}
	fmt.Fprintf(w, "%s", outputJson)
}

func getClinicSvcQueueMetaHandler2(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	r.ParseForm()

	//TODO: This line is an ad-hoc solution to remove the 
	//      std output triggered by CORS preflighted request via http OPTIONS.
	//      The correct solution is to add a reverse-proxy server to the
	//      main server so that all requests are channeled to this reverse proxy
	//      in order to overcome CORS restriction.
	if (len(r.PostForm) == 0) { return } 

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()	
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	defer func() {
		if err=client.Disconnect(ctx); err!=nil {
			panic(err)
		}
	}()
	mytcaDB := client.Database("test")
	bookingColl := mytcaDB.Collection("booking3")

	queuesCapPerHrProjPath := "monthlyOpSchedule"
	projection := bson.D{
		{queuesCapPerHrProjPath, 
			bson.D{
				{"$elemMatch", bson.D{{"date", r.PostForm["date"][0]}}},
			},
		},		
	}

	opts := options.FindOne()
	opts.SetProjection(projection)
	var clinicSvcMongodbMeta ClinicSvcMongodbMeta
	// var clinicSvcMongodbQueueMeta ClinicSvcMongodbQueueMeta
	var clinicSvcMongodbQueueMeta DailyOpSchedule3
	err = bookingColl.FindOne(
		ctx,
		bson.D{
			{"clinic" , r.PostForm["clinicId"][0]},
			{"dept", "opd_disease"}, //TODO: to map 'service' to 'dept' value.
			{"monthlyOpSchedule.date", r.PostForm["date"][0]},
		},
		opts,
	).Decode(&clinicSvcMongodbMeta)
	if err != nil {
		log.Print(err)
		w.WriteHeader(500)
		fmt.Fprintf(w, "Internal server error!")
		return
	}
	clinicSvcMongodbQueueMeta = clinicSvcMongodbMeta.MonthlyOpSchedules[0]
	// fmt.Println("ClinicSvcMongodbQueueMeta: \n", clinicSvcMongodbQueueMeta)

	outputJson, err := json.MarshalIndent(&clinicSvcMongodbQueueMeta, "", "\t")
	if err != nil {
		log.Print(err)
		w.WriteHeader(500)
		fmt.Fprintf(w, "Internal server error!")
		return
	}
	fmt.Fprintf(w, "%s", outputJson)
}


func main() {
	// USER 
	// ====
	// http.HandleFunc("/", helloWorldPingMongodbHandler)
	// http.HandleFunc("/booking", getDailyOpSchedule)
	// http.HandleFunc("/submit", submitHandler)
	// http.HandleFunc("/allclinics", getClinicsHandler)
	// http.HandleFunc("/checkservice", checkIfClinicSvcExistHandler)
	// http.HandleFunc("/serviceophours", getClinicSvcMetaHandler)
	// http.HandleFunc("/serviceophours", getClinicSvcMetaHandler2)
	// http.HandleFunc("/servicebookings", getClinicSvcQueueMetaHandler)
	// http.HandleFunc("/servicebookings", getClinicSvcQueueMetaHandler2)
	// http.HandleFunc("/book", makeBookingHandler)
	// http.HandleFunc("/book", makeBookingHandler2)
	// http.HandleFunc("/book", makeBookingHandler3)
	http.HandleFunc("/allclinics", my_ldap.GetAllClinicsHandler)
	http.HandleFunc("/checkservice", my_ldap.CheckIfSvcExistHandler)
	http.HandleFunc("/serviceophours", my_ldap.GetSvcOpHrsHandler)
	http.HandleFunc("/servicebookings", my_mongodb.GetSvcBookingsHandler)
	http.HandleFunc("/book", my_mongodb.MakeBookingHandler)
	http.HandleFunc("/patientbookings", my_mongodb.GetPatientBookingsHandler)
	http.HandleFunc("/patientdelbookings", my_mongodb.DelBookingHandler)

	// ADMIN
	// =====
	// Pkd
	http.HandleFunc("/admin/pkdclinics", my_ldap.GetPkdClinicsHandler)
	// Clinic
	http.HandleFunc("/admin/clinicdetails", GetClinicDetailsHandler)
	http.HandleFunc("/admin/updateclinic", updateClinicHandler)
	// Dept
	http.HandleFunc("/admin/clinicdeptdetails", GetClinicDeptDetailsHandler)
	http.HandleFunc("/admin/adddept", addDeptHandler)
	http.HandleFunc("/admin/updatedept", updateDeptHandler)
	http.HandleFunc("/admin/toggledeptavai", toggleDeptAvaiHandler)
	// Service
	http.HandleFunc("/admin/clinicsvcbscdetails", GetClinicServiceBasicDetailsHandler)	
	http.HandleFunc("/admin/clinicsvcadvdetails", GetClinicServiceAdvDetailsHandler)
	http.HandleFunc("/admin/addsvc", addSvcHandler)
	http.HandleFunc("/admin/updatesvc", updateSvcHandler)
	http.HandleFunc("/admin/togglesvcavai", toggleSvcAvaiHandler)
	
	// AUTH
	// ====
	http.HandleFunc("/public/login", auth.BindHandler)
	http.HandleFunc("/public/signup", auth.CreateAccountHandler)
	
	// SERVER
	// ======
	http.ListenAndServe(":8082", nil)


	// INIT MONGODB DATABASE 
	// =====================
	// InitOpSchedule3(2020, 10, "opd", "kk_maran", "maran", "pahang")
	// InitOpSchedule4(2020, 10, "kk_maran", "maran", "pahang")
	// InitOpSchedule5(2020, 10, "kk_maran", "maran", "pahang")
	// my_mongodb.InitOpSchedule("880601105150", "88motherfaker88", 2020, 10, "kk_maran", "maran", "pahang")


	// DEBUG OUTPUT 
	// ============
	// #DEBUG-1
	// clinicsDirJson, err := GetAllClinics()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("Clinic Directory: \n%s\n", clinicsDirJson)

	// #DEBUG-2
	// m, err := GetClinicDeptAndServicesMeta("kk_maran", "maran", "pahang")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println("ClinicDeptAndSvcMeta: \n", m)

	// #DEBUG-3
	// meta, err := GetClinicSvcOpHrs("diabetes", "opd_disease", "kk_maran", "maran", "pahang")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println("GetClinicSvcOpHrs: \n", meta)

	// #DEBUG-4
	// ifSvcExist, err := checkIfSvcExist("diabetes", "opd_disease", "kk_maran", "maran", "pahang")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println("ifSvcExist: ", ifSvcExist)	

	// #DEBUG-5
	// checkIfClinicSvcExistHandler()

	// #DEBUG-6
	// _, err := getClinicDetails("880601105150", "88motherfaker88", "kk_maran", "maran", "pahang")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	
	// #DEBUG-7
    // _, err := getClinicDeptDetails("880601105150", "88motherfaker88", "opd_disease", "kk_maran", "maran", "pahang")
    // _, err := getClinicDeptDetails("880601105150", "88motherfaker88", "abc", "kk_maran", "maran", "pahang")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// #DEBUG-8
	// _, err := getClinicServiceAdvDetails(true, true, "880601105150", "88motherfaker88","x-ray", "x-ray", "kk_maran", "maran", "pahang")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// #DEBUG-9
	// _, err := getClinicServiceBasicDetails("880601105150", "88motherfaker88", "opd_disease", "kk_maran", "maran", "pahang")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	
	// #DEBUG-10
	// deptDatalist, err := db.GetDeptNameAndStaffNum("880601105150", "88motherfaker88", "kk_maran", "maran", "pahang")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("deptDatalist: %+v \n", deptDatalist)

	// #DEBUG-11
	// svcOpHrs, err := my_ldap.GetSvcOpHrs("880601105150", "88motherfaker88", "diabetes", "opd_disease", "kk_maran", "maran", "pahang")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("svcOpHrs: %+v \n", svcOpHrs)

	// #DEBUG-12
	// err := my_ldap.Bind("880601105149", "88motherfaker")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// #DEBUG-13
	// tokenString, err := auth.NewTokenHMAC("880601105149")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("tokenString: %s \n", tokenString)

	// tokenValid := auth.VerifyTokenHMAC(tokenString)
	// fmt.Printf("Is token valid: %v \n", tokenValid)
}
	