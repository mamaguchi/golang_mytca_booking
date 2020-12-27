package main

import (
	"go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
	"fmt"
	"time"
	"context"
	"log"
	"strconv"
	"strings"
	// "github.com/go-ldap/ldap/v3"
    // "go.mongodb.org/mongo-driver/mongo/readpref"
	//"net/http"
	//"encoding/json"
)

// const (
// 	DIR_MGR_DN       =	"cn=Directory Manager"
// 	DIR_MGR_PWD      =	"88motherfaker88"
// 	STAFF_BASE_DN    =	"ou=people,dc=example,dc=com"
// 	CLINIC_BASE_DN   =  "ou=kkm-clinic,ou=groups,dc=example,dc=com"
// 	SERVICE_TEMPLATE_DN =  "clinicServiceName=%s,ou=service,cn=%s,ou=pkd_%s,ou=jkn_%s," + CLINIC_BASE_DN
// )

// type Booking struct {
// 	State string 			`bson:"state" json:"state"`
// 	District string 		`bson:"district" json:"district"`
// 	Clinic string 			`bson:"clinic" json:"clinic"` 
// 	Service string 			`bson:"service" json:"service"` 
// 	CloseDays []int 		`bson:"closeDays" json:"closeDays"`
// 	HalfDays []int 			`bson:"halfDays" json:"halfDays"`
// 	StartOpHr int 			`bson:"startOpHr" json:"startOpHr"`
// 	EndOpHr int 			`bson:"EndOpHr" json:"EndOpHr"`
// 	StartOpHrHalfDay int 	`bson:"startOpHrHalfDay" json:"startOpHrHalfDay"`
// 	EndOpHrHalfDay int 		`bson:"EndOpHrHalfDay" json:"EndOpHrHalfDay"`
// 	PublicHolMonth []int 	`bson:"publicHolMonth" json:"publicHolMonth"`
// 	StaffDaily int 			`bson:"staffDaily" json:"staffDaily"` 
// 	AvgConsultTime int 		`bson:"avgConsultTime" json:"avgConsultTime"` 
// 	MonthlyOpSchedule []DailyOpSchedule`bson:"monthlyOpSchedule" json:"monthlyOpSchedule"` 
// }

type DailyOpSchedule struct {
	Date string 				`bson:"date" json:"date"`
	DayOfWeek int				`bson:"dayOfWeek" json:"dayOfWeek"`
	StaffPerDay []int 			`bson:"staffPerDay" json:"staffPerDay"`
	QueuesCapPerDay []int 		`bson:"queuesCapPerDay" json:"queuesCapPerDay"` 
	QueuesUsgPerDay []int 		`bson:"queuesUsgPerDay" json:"queuesUsgPerDay"` 
	QueuesPerDay []QueuePerHr 	`bson:"queuesPerDay" json:"queuesPerDay"` 
}

type DailyOpSchedule4 struct {
	Date string 				`bson:"date" json:"date"`
	DayOfWeek int				`bson:"dayOfWeek" json:"dayOfWeek"`
	QueuesCapPerDay []int 		`bson:"queuesCapPerDay" json:"queuesCapPerDay"` 
	QueuesUsgPerDay []int 		`bson:"queuesUsgPerDay" json:"queuesUsgPerDay"` 
	QueuesPerDay []QueuePerHr	`bson:"queuesPerDay" json:"queuesPerDay"` 
}

type QueuePerHr struct {
	PatientIds []int 			`bson:"patientIds" json:"patientIds"`
	BookingReasons []int 		`bson:"bookingReasons" json:"bookingReasons"`
}


func InitOpSchedule() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
    defer func() {
		if err=client.Disconnect(ctx); err!=nil {
			panic(err)
		}
		}()
		
	mytcaDB := client.Database("test")
	bookingColl := mytcaDB.Collection("booking")

	queuePerHr := QueuePerHr{
							PatientIds: []int{}, 
							BookingReasons: []int{},
				}

	dailyOpSchedule := DailyOpSchedule{
							Date: "2020-08-28",
							// IsHalfDay: 1,
							StaffPerDay: []int{4, 2},
							QueuesCapPerDay: []int{36, 36},
							QueuesPerDay: []QueuePerHr{queuePerHr},
				}

    res, err := bookingColl.UpdateOne(
		ctx, 
		bson.M{"clinic" : "kk_maran"},
		bson.D{
			{"$push", bson.D{{"monthlyOpSchedule", dailyOpSchedule}}},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Updated %v MongoDB Documents\n", res.ModifiedCount)
}

func InitOpSchedule2() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
    defer func() {
		if err=client.Disconnect(ctx); err!=nil {
			panic(err)
		}
		}()
		
	mytcaDB := client.Database("test")
	bookingColl := mytcaDB.Collection("booking")

	staff := 5
	avgConsultTime := 10

	year := 2020
	var month time.Month = 8
	t := time.Date(year, month, 1, 0,0,0,0,time.UTC)
	lastDayOfMonth := time.Date(year, month+1, 0, 0,0,0,0,time.UTC).Day()
	monthlyOpSchedule := []DailyOpSchedule{}

	//24-hour format
	startOpHr, endOpHr := 8, 17
	startOpHrHalfDay, endOpHrHalfDay := 8, 13
	
	for i:=1; i<=lastDayOfMonth; i++ {
		if(t.Weekday()==0) {
			//It's close day, so do nothing.
		} else if(t.Weekday()==6) {
			//It's a half-day
			staffPerDay := []int{} 
			queuesCapPerDay := []int{} 
			queuesPerDay := []QueuePerHr{}
			
			for j:=startOpHrHalfDay; j < endOpHrHalfDay; j++ {
				queueCapPerHr := staff * 60 / avgConsultTime
				queuePerHr := QueuePerHr{
					PatientIds: []int{}, 
					BookingReasons: []int{},
					}

				staffPerDay = append(staffPerDay, staff)
				queuesCapPerDay = append(queuesCapPerDay, queueCapPerHr)
				queuesPerDay = append(queuesPerDay, queuePerHr)
			}

			dailyOpSchedule := DailyOpSchedule{
				Date: t.String()[:10],
				// IsHalfDay: 1,
				StaffPerDay: staffPerDay,
				QueuesCapPerDay: queuesCapPerDay,
				QueuesPerDay: queuesPerDay,
				}
			monthlyOpSchedule = append(monthlyOpSchedule, dailyOpSchedule)
		} else {
			//It's a full-day
			staffPerDay := []int{} 
			queuesCapPerDay := []int{} 
			queuesPerDay := []QueuePerHr{}
			
			for j:=startOpHr; j < endOpHr; j++ {
				queueCapPerHr := staff * 60 / avgConsultTime
				queuePerHr := QueuePerHr{
					PatientIds: []int{}, 
					BookingReasons: []int{},
					}

				staffPerDay = append(staffPerDay, staff)
				queuesCapPerDay = append(queuesCapPerDay, queueCapPerHr)
				queuesPerDay = append(queuesPerDay, queuePerHr)
			}

			dailyOpSchedule := DailyOpSchedule{
				Date: t.String()[:10],
				// IsHalfDay: 0,
				StaffPerDay: staffPerDay,
				QueuesCapPerDay: queuesCapPerDay,
				QueuesPerDay: queuesPerDay,
				}
			monthlyOpSchedule = append(monthlyOpSchedule, dailyOpSchedule)
		}

		t = t.AddDate(0, 0, 1)
	}

    res, err := bookingColl.UpdateOne(
		ctx, 
		bson.M{"clinic" : "kk_maran"},
		bson.D{
			{"$set", bson.D{{"monthlyOpSchedule", monthlyOpSchedule}}},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Updated %v MongoDB Documents\n", res.ModifiedCount)
}

func InitOpSchedule3(year int, month int, service string,
					clinic string, district string, state string) (err error){
	meta, err := GetClinicServiceMeta(service, clinic, district, state)
	if err != nil {
		return
	}

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

	var mth time.Month = time.Month(month)
	t := time.Date(year, mth, 1, 0,0,0,0,time.UTC)
	lastDayOfMonth := time.Date(year, time.Month(month+1), 0, 0,0,0,0,time.UTC).Day()
	monthlyOpSchedule := []DailyOpSchedule{}

	startOpHrsStr := strings.Split(meta.StartOpHrs, ",")    //24-hour format
	endOpHrsStr := strings.Split(meta.EndOpHrs, ",")		//24-hour format
	avaiDaysStr := strings.Split(meta.AvaiDays, ",")
	var avaiDays [7]int
	for i,_ := range avaiDaysStr {
		avaiDays[i], _ = strconv.Atoi(avaiDaysStr[i])
	}
	
	for i:=1; i<=lastDayOfMonth; i++ {
		if(avaiDays[t.Weekday()]==1) {
			//It's a working day
			startOpHr, _ := strconv.Atoi(startOpHrsStr[t.Weekday()])
			endOpHr, _ := strconv.Atoi(endOpHrsStr[t.Weekday()])
			staffPerDay := []int{} 
			queuesCapPerDay := []int{} 
			queuesUsgPerDay := []int{} 
			queuesPerDay := []QueuePerHr{}
			queueCapPerHr := meta.NumOfStaff * 60
			queueUsgPerHr := 0
			queuePerHr := QueuePerHr{
					PatientIds: []int{}, 
					BookingReasons: []int{},
				}
			
			for j:=startOpHr; j < endOpHr; j++ {
				staffPerDay = append(staffPerDay, meta.NumOfStaff)
				queuesCapPerDay = append(queuesCapPerDay, queueCapPerHr)
				queuesUsgPerDay = append(queuesUsgPerDay, queueUsgPerHr)
				queuesPerDay = append(queuesPerDay, queuePerHr)
			}

			dailyOpSchedule := DailyOpSchedule{
				Date: t.String()[:10],
				DayOfWeek: int(t.Weekday()),
				StaffPerDay: staffPerDay,
				QueuesCapPerDay: queuesCapPerDay,
				QueuesUsgPerDay: queuesUsgPerDay,
				QueuesPerDay: queuesPerDay,
				}
			monthlyOpSchedule = append(monthlyOpSchedule, dailyOpSchedule)
		}

		t = t.AddDate(0, 0, 1)
	}

    res, err := bookingColl.InsertOne(
		ctx, 
		bson.D{
			{"clinic", clinic},
			{"service", service},
			{"monthlyOpSchedule", monthlyOpSchedule},
		},
	)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("The _id field of the inserted document, or nil if no insert was done: %v\n", res.InsertedID)

	return
}

func aggregateClinicSvcAvaiDays(clinicSvcAvaiDays []string) ([7]int) {
	type bits uint8
	var flag bits
	for _, str := range clinicSvcAvaiDays {
	    strSplit := strings.Split(str, ",")
	    for i:=0; i<len(strSplit); i++ {
            var val bits
            tmp, _ := strconv.Atoi(strSplit[i])
            val = bits(tmp)
            flag = flag | val << i
	    }	
	}
	// fmt.Printf("Avai Days Flag(reversed): %#b \n", flag)
	
	var aggrAvaiDays [7]int
	for k:=0; k<7; k++ {
	    aggrAvaiDays[k] = int(1 & (flag >> k))
	}
	return aggrAvaiDays
}

func InitOpSchedule4(year int, month int, clinic string, district string, state string) (err error){

	metaMap, err := GetClinicDeptAndServicesMeta(clinic, district, state)
	if err != nil {
		return
	}

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

	for _, deptAndSvcMetas := range metaMap {
		var deptName string 
		var deptNumOfStaff int
		var deptStartHrs, deptEndHrs string
		var svcAvaiDaysStrArr []string
		for _, deptAndSvcMeta := range deptAndSvcMetas {
			if deptAndSvcMeta.Type == "dept" {
				deptName = deptAndSvcMeta.DeptName 
				deptNumOfStaff = deptAndSvcMeta.DeptNumOfStaff
				deptStartHrs = deptAndSvcMeta.DeptStartHrs 
				deptEndHrs = deptAndSvcMeta.DeptEndHrs 
			} else if deptAndSvcMeta.Type == "svc" {
				svcAvaiDaysStrArr = append(svcAvaiDaysStrArr, deptAndSvcMeta.SvcAvaiDays)
			}
		}

		var mth time.Month = time.Month(month)
		t := time.Date(year, mth, 1, 0,0,0,0,time.UTC)
		lastDayOfMonth := time.Date(year, time.Month(month+1), 0, 0,0,0,0,time.UTC).Day()
		monthlyOpSchedule := []DailyOpSchedule4{}
		
		startOpHrsStr := strings.Split(deptStartHrs, ",")    //24-hour format
		endOpHrsStr := strings.Split(deptEndHrs, ",")		 //24-hour format
		avaiDays := aggregateClinicSvcAvaiDays(svcAvaiDaysStrArr)
		
		for i:=1; i<=lastDayOfMonth; i++ {
			if(avaiDays[t.Weekday()]==1) {    //It's a working day				
				startOpHr, _ := strconv.Atoi(startOpHrsStr[t.Weekday()])
				endOpHr, _ := strconv.Atoi(endOpHrsStr[t.Weekday()])
				queuesCapPerDay := []int{} 
				queuesUsgPerDay := []int{} 
				queuesPerDay := []QueuePerHr{}
				queueCapPerHr := deptNumOfStaff * 60
				queueUsgPerHr := 0
				queuePerHr := QueuePerHr{
					PatientIds: []int{}, 
					BookingReasons: []int{},
				}
	
				for j:=startOpHr; j < endOpHr; j++ {
					queuesCapPerDay = append(queuesCapPerDay, queueCapPerHr)
					queuesUsgPerDay = append(queuesUsgPerDay, queueUsgPerHr)
					queuesPerDay = append(queuesPerDay, queuePerHr)
				}
	
				dailyOpSchedule := DailyOpSchedule4{
					Date: t.String()[:10],
					DayOfWeek: int(t.Weekday()),
					QueuesCapPerDay: queuesCapPerDay,
					QueuesUsgPerDay: queuesUsgPerDay,
					QueuesPerDay: queuesPerDay,
				}
				monthlyOpSchedule = append(monthlyOpSchedule, dailyOpSchedule)
			}	
			t = t.AddDate(0, 0, 1)
		}

		res, err := bookingColl.InsertOne(
			ctx, 
			bson.D{
				{"clinic", clinic},
				{"dept", deptName},
				{"monthlyOpSchedule", monthlyOpSchedule},
			},
		)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("The _id field of the inserted document, or nil if no insert was done: %v\n", res.InsertedID)
	}

	return
}

type DailyOpSchedule5 struct {
	Date string 				`bson:"date" json:"date"`
	DayOfWeek int				`bson:"dayOfWeek" json:"dayOfWeek"`
	QueuesCapPerDay []int 		`bson:"queuesCapPerDay" json:"queuesCapPerDay"` 
	QueuesUsgPerDay []int 		`bson:"queuesUsgPerDay" json:"queuesUsgPerDay"` 
	QueuesPerDay []QueueByHr   `bson:"queuesPerDay" json:"queuesPerDay"` 
}

type QueueByHr struct {
	Bookings []Booking 			`bson:"bookings" json:"bookings"`	
}

type Booking struct {
	PatientId int 		   		`bson:"patientId" json:"patientId"`
	BookingReason int 			`bson:"bookingReason" json:"bookingReason"`
}
	
func InitOpSchedule5(year int, month int, clinicId string, district string, state string) (err error){

	metaMap, err := GetClinicDeptAndServicesMeta(clinicId, district, state)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()	
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	defer func() {
		if err=client.Disconnect(ctx); err!=nil {
			panic(err)
		}
	}()
	mytcaDB := client.Database("test")
	bookingColl := mytcaDB.Collection("booking4")

	for _, deptAndSvcMetas := range metaMap {
		var deptName string 
		var deptNumOfStaff int
		// var deptStartHrs, deptEndHrs string
		// var svcAvaiDaysStrArr []string
		for _, deptAndSvcMeta := range deptAndSvcMetas {
			if deptAndSvcMeta.Type == "dept" {
				deptName = deptAndSvcMeta.DeptName 
				deptNumOfStaff = deptAndSvcMeta.DeptNumOfStaff
				// deptStartHrs = deptAndSvcMeta.DeptStartHrs 
				// deptEndHrs = deptAndSvcMeta.DeptEndHrs 
			}
			// } else if deptAndSvcMeta.Type == "svc" {
			// 	svcAvaiDaysStrArr = append(svcAvaiDaysStrArr, deptAndSvcMeta.SvcAvaiDays)
			// }
		}

		var mth time.Month = time.Month(month)
		t := time.Date(year, mth, 1, 0,0,0,0,time.UTC)
		lastDayOfMonth := time.Date(year, time.Month(month+1), 0, 0,0,0,0,time.UTC).Day()
		monthlyOpSchedule := []DailyOpSchedule5{}
		
		// startHrsStr := strings.Split(deptStartHrs, ",")  //24-hour format
		// endHrsStr := strings.Split(deptEndHrs, ",")		 //24-hour format
		// avaiDays := aggregateClinicSvcAvaiDays(svcAvaiDaysStrArr)
		
		for i:=1; i<=lastDayOfMonth; i++ {
			// if(avaiDays[t.Weekday()]==1) { //It's a working day				
			// startOpHr, _ := strconv.Atoi(startHrsStr[t.Weekday()])
			// endOpHr, _ := strconv.Atoi(endHrsStr[t.Weekday()])
			queuesCapPerDay := []int{} 
			queuesUsgPerDay := []int{} 
			queuesPerDay := []QueueByHr{}
			queueCapPerHr := deptNumOfStaff * 60
			queueUsgPerHr := 0
			queueByHr := QueueByHr{
				Bookings: []Booking{},
			}

			for j:=0; j < 24; j++ {
				queuesCapPerDay = append(queuesCapPerDay, queueCapPerHr)
				queuesUsgPerDay = append(queuesUsgPerDay, queueUsgPerHr)
				queuesPerDay = append(queuesPerDay, queueByHr)
			}

			dailyOpSchedule := DailyOpSchedule5{
				Date: t.String()[:10],
				DayOfWeek: int(t.Weekday()),
				QueuesCapPerDay: queuesCapPerDay,
				QueuesUsgPerDay: queuesUsgPerDay,
				QueuesPerDay: queuesPerDay,
			}
			monthlyOpSchedule = append(monthlyOpSchedule, dailyOpSchedule)
			// }	
			t = t.AddDate(0, 0, 1)
		}

		entryDate := strconv.Itoa(year) + "-" + strconv.Itoa(month)

		res, err := bookingColl.InsertOne(
			ctx, 
			bson.D{
				{"clinicId", clinicId},
				{"dept", deptName},
				{"date", entryDate},
				{"monthlyOpSchedule", monthlyOpSchedule},
			},
		)
		if err != nil {
			log.Print(err)
			return err
		}
		fmt.Printf("The _id field of the inserted document (nil if no insert was done):\t %v\n", res.InsertedID)
	}

	return
}