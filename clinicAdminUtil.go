package main

import (
	"log"
	"fmt"
	"github.com/go-ldap/ldap/v3"
	// "github.com/google/uuid"
	// "strings"
	"strconv"
	"encoding/json"
	"net/http"
)

type ClinicDetails struct {
	ClinicName string 				`json:"clinicName"`
	ClinicState string 				`json:"clinicState"`
	ClinicDistrict string			`json:"clinicDistrict"`
	ClinicCloseDays string 			`json:"clinicCloseDays"`
	ClinicHalfDays string 			`json:"clinicHalfDays"`
	ClinicPublicHolByMonth []string `json:"clinicPublicHolidays"`
	ClinicStaffIds []string 		`json:"clinicStaffIds"`
	ClinicDeptBasicDetailsLst []ClinicDeptBasicDetails  `json:"clinicDeptBasicDetailsLst"`
}

type ClinicDeptBasicDetails struct {
	ClinicDeptName string			`json:"clinicDeptName"`
	ClinicDeptIsEnabled int 		`json:"clinicDeptIsEnabled"`
}

type ClinicDeptAdvDetails struct {
	ClinicDeptAvaiDays string		`json:"clinicDeptAvaiDays"`
	ClinicDeptStartHrs string		`json:"clinicDeptStartHrs"`
	ClinicDeptEndHrs string			`json:"clinicDeptEndHrs"`
	ClinicDeptNumOfStaff int		`json:"clinicDeptNumOfStaff"`
	ClinicDeptStaffIds []string		`json:"clinicDeptStaffIds"`
	ClinicDeptMaxPt int				`json:"clinicDeptMaxPt"`
}

type ServiceBasicDetailsList struct {
	ServiceBasicDetailsLst []ServiceBasicDetails `json:"serviceBasicDetailsLst"`
}

type ServiceBasicDetails struct {
	Name string		 				`json:"name"`
	IsEnabled int 					`json:"isEnabled"`	
}

type ServiceAdvDetails struct {
	AvaiDays string 				`json:"avaiDays"`
	StartHrs string 				`json:"startHrs"`
	EndHrs string 					`json:"endHrs"`
	AvgConsultTime int 				`json:"avgConsultTime"`
}

func getClinicDetails(userId string, userPwd string, clinicId string, district string, state string) (outputJson []byte, err error) {
	l, err := ldap.DialURL("ldap://127.0.0.1:389")
	if err != nil {
		log.Print(err)
		return nil, err
	}
	defer l.Close()

	// TODO: KIV to change 'staffId' to 'userId' by changing
	//       the LDAP object class from 'staff' to regular 'user'
	userDN := fmt.Sprintf("staffId=%s,%s", userId, STAFF_BASE_DN)
	err = l.Bind(userDN, userPwd)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	// Search for the clinic 
	clinicDN := fmt.Sprintf(CLINIC_TEMPLATE_DN, clinicId, district, state)
	searchRequest := ldap.NewSearchRequest(
		clinicDN,
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0,0, false,
		"(&)",
		[]string{"*"},
		nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
			log.Print(err)
			return nil, err
	}
	clinicName := sr.Entries[0].GetAttributeValue("clinicName")
	clinicState := sr.Entries[0].GetAttributeValue("clinicState")
	clinicDistrict := sr.Entries[0].GetAttributeValue("clinicDistrict")
	clinicCloseDays:= sr.Entries[0].GetAttributeValue("clinicCloseDays")
	clinicHalfDays := sr.Entries[0].GetAttributeValue("clinicHalfDays")
	publicHolByMonth := sr.Entries[0].GetAttributeValues("publicHolByMonth")
	clinicStaffIds := sr.Entries[0].GetAttributeValues("staffId")

	clinicDetails := ClinicDetails{
		ClinicName: clinicName,
		ClinicState: clinicState,
		ClinicDistrict: clinicDistrict,
		ClinicCloseDays: clinicCloseDays,
		ClinicHalfDays: clinicHalfDays,
		ClinicPublicHolByMonth: publicHolByMonth,
		ClinicStaffIds: clinicStaffIds,
	}
	
	// Search for the departments of the clinic
	deptBaseDN := fmt.Sprintf(DEPT_BASE_TEMPLATE_DN, clinicId, district, state)
	searchFilter := "(clinicDeptName=*)"
	searchRequest = ldap.NewSearchRequest(
			deptBaseDN,
			ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0,0, false,
			searchFilter,
			[]string{"clinicDeptName", "clinicDeptIsEnabled"},
			nil,
	)
	sr, err = l.Search(searchRequest)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	var deptBasicDetailsLst []ClinicDeptBasicDetails
	for _, entry := range sr.Entries {
		clinicDeptName := entry.GetAttributeValue("clinicDeptName")
		clinicDeptIsEnabled,_ := strconv.Atoi(entry.GetAttributeValue("clinicDeptIsEnabled"))
		
		deptBasicDetails := ClinicDeptBasicDetails{
			ClinicDeptName: clinicDeptName,
			ClinicDeptIsEnabled: clinicDeptIsEnabled,
		}
		deptBasicDetailsLst = append(deptBasicDetailsLst, deptBasicDetails)
	}
	clinicDetails.ClinicDeptBasicDetailsLst = deptBasicDetailsLst
	// fmt.Printf("Clinic Details :\n%+v\n", clinicDetails)

	outputJson, err = json.MarshalIndent(clinicDetails, "", "\t")
	if err != nil {
		log.Print(err)
		return nil, err
	}
	fmt.Printf("\nClinic Details Response[JSON]:\n%s\n", outputJson)

	return 
}

func GetClinicDetailsHandler(w http.ResponseWriter, r *http.Request) {	
	w.Header().Set("Access-Control-Allow-Origin", "*")
	r.ParseForm()
	fmt.Println("[GetClinicDetailsHandler] Request Form Data Received! \n")
	fmt.Println(r.Form)
	
	userId := r.Form["userId"][0]
	userPwd := r.Form["userPwd"][0]
	clinicId := r.Form["clinicId"][0]
	district := r.Form["district"][0]
	state := r.Form["state"][0]	
	
	clinicDetails, err := getClinicDetails(userId, userPwd, clinicId, district, state)
	if err != nil {
		log.Print(err)
		w.WriteHeader(500)
		fmt.Fprintf(w, "Internal server error!")
		return
	}
	fmt.Fprintf(w, "%s", clinicDetails)
}

func getClinicDeptDetails(isDirty bool, isEnabled bool, userId string, userPwd string, deptName string,
						 clinicId string, district string, state string) (outputJson []byte, err error) {
	l, err := ldap.DialURL("ldap://127.0.0.1:389")
	if err != nil {
		log.Print(err)
		return nil, err
	}
	defer l.Close()

	// TODO: KIV to change 'staffId' to 'userId' by changing
	//       the LDAP object class from 'staff' to regular 'user'
	userDN := fmt.Sprintf("staffId=%s,%s", userId, STAFF_BASE_DN)
	err = l.Bind(userDN, userPwd)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	deptBaseDN := fmt.Sprintf(DEPT_TEMPLATE_DN, deptName, clinicId, district, state)

	// If 'isDirty' is true, dept 'isEnabled' status
	// has been changed by client app.
	// Thus, we need to first update the dept 'isEnabled'
	// attribute value before retrieving dept advanced data,
	// and sending them to client app.
	if isDirty {
		modifyRequest := ldap.NewModifyRequest(
			deptBaseDN, 
			nil,
		)
		var isEnabledStr string
		if isEnabled {
			isEnabledStr = "1"
		} else {
			isEnabledStr = "0"
		}
		modifyRequest.Replace("clinicDeptIsEnabled", []string{isEnabledStr})
	
		err = l.Modify(modifyRequest)
		if err != nil {
				log.Print(err)
				return nil, err
		}
	} 
	
	// Search for the advanced data for the dept of the clinic
	searchFilter := "(&)"	
	searchRequest := ldap.NewSearchRequest(
			deptBaseDN,
			ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0,0, false,
			searchFilter,
			[]string{"*"},
			nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Print(err)
		resultCode := err.(*ldap.Error).ResultCode //uint16 
		if resultCode == 32 {
			//LDAPResultNoSuchObject
			fmt.Println("There is no such Department object found!")
		}	
		return nil, err
	}

	entry := sr.Entries[0]
	clinicDeptAvaiDays := entry.GetAttributeValue("clinicDeptAvaiDays")
	clinicDeptStartHrs := entry.GetAttributeValue("clinicDeptStartHour")
	clinicDeptEndHrs := entry.GetAttributeValue("clinicDeptEndHour")
	clinicDeptNumOfStaff, _ := strconv.Atoi(entry.GetAttributeValue("clinicDeptNumOfStaff"))
	clinicDeptStaffIds := entry.GetAttributeValues("clinicDeptStaffId")
	clinicDeptMaxPt, _ := strconv.Atoi(entry.GetAttributeValue("clinicDeptMaxPt"))

	clinicDeptAdvDetails := ClinicDeptAdvDetails{
		ClinicDeptAvaiDays: clinicDeptAvaiDays,
		ClinicDeptStartHrs: clinicDeptStartHrs,
		ClinicDeptEndHrs: clinicDeptEndHrs,
		ClinicDeptNumOfStaff: clinicDeptNumOfStaff,
		ClinicDeptStaffIds: clinicDeptStaffIds,
		ClinicDeptMaxPt: clinicDeptMaxPt,
	}	

	outputJson, err = json.MarshalIndent(clinicDeptAdvDetails, "", "\t")
	if err != nil {
		log.Print(err)
		return nil, err
	}
	// fmt.Printf("\nClinic Dept Details [JSON]:\n%s\n", outputJson)

	return 
}

func GetClinicDeptDetailsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")	

	//TODO: This line is an ad-hoc solution to remove the 
	//      std output triggered by CORS preflighted request via http OPTIONS.
	//      The correct solution is to add a reverse-proxy server to the
	//      main server so that all requests are channeled to this reverse proxy
	//      in order to overcome CORS restriction.
	if (r.Method == "OPTIONS") { return } 
	
	r.ParseForm()
	fmt.Println("[GetClinicDeptDetailsHandler] Request Form Data Received! \n")
	fmt.Println(r.Form)
	
	isDirty, _ := strconv.ParseBool(r.Form["isDirty"][0])
	isEnabled, _ := strconv.ParseBool(r.Form["isEnabled"][0])
	userId := r.Form["userId"][0]
	userPwd := r.Form["userPwd"][0]
	deptName := r.Form["deptName"][0]
	clinicId := r.Form["clinicId"][0]
	district := r.Form["district"][0]
	state := r.Form["state"][0]
	
	clinicDeptAdvDetails, err := getClinicDeptDetails(isDirty, isEnabled, userId, userPwd, 
												deptName, clinicId, district, state)
	if err != nil {
		log.Print(err)
		w.WriteHeader(500)
		fmt.Fprintf(w, "Internal server error!")
		return
	}
	fmt.Printf("\nDept Details Response [JSON]:\n%s\n", clinicDeptAdvDetails)
	fmt.Fprintf(w, "%s", clinicDeptAdvDetails)
}

func getClinicServiceBasicDetails(userId string, userPwd string,
	deptName string, clinicId string, district string, state string) (outputJson []byte, err error) {
	
	l, err := ldap.DialURL("ldap://127.0.0.1:389")
	if err != nil {
		log.Print(err)
		return nil, err
	}
	defer l.Close()

	// TODO: KIV to change 'staffId' to 'userId' by changing
	//       the LDAP object class from 'staff' to regular 'user'
	userDN := fmt.Sprintf("staffId=%s,%s", userId, STAFF_BASE_DN)
	err = l.Bind(userDN, userPwd)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	serviceBaseDN := fmt.Sprintf(SERVICE_BASE_TEMPLATE_DN, deptName, clinicId, district, state)

	// Search for the advanced data for the dept of the clinic
	searchFilter := "(&)"	
	searchRequest := ldap.NewSearchRequest(
		serviceBaseDN,
		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0,0, false,
		searchFilter,
		[]string{"clinicServiceName", "clinicServiceIsEnabled"},
		nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Print(err)
		resultCode := err.(*ldap.Error).ResultCode //uint16 
		if resultCode == 32 {
			//LDAPResultNoSuchObject
			fmt.Println("There is no such Service object found!")
		}	
		return nil, err
	}

	var svcBasicDetailsLst []ServiceBasicDetails
	for _, entry := range sr.Entries {
		clinicServiceName := entry.GetAttributeValue("clinicServiceName")
		clinicServiceIsEnabled,_ := strconv.Atoi(entry.GetAttributeValue("clinicServiceIsEnabled"))
		
		svcBasicDetails := ServiceBasicDetails{
			Name: clinicServiceName,
			IsEnabled: clinicServiceIsEnabled,
		}
		svcBasicDetailsLst = append(svcBasicDetailsLst, svcBasicDetails)
	}
	serviceBasicDetailsList := ServiceBasicDetailsList{
		ServiceBasicDetailsLst: svcBasicDetailsLst,
	}

	outputJson, err = json.MarshalIndent(serviceBasicDetailsList, "", "\t")
	if err != nil {
		log.Print(err)
		return nil, err
	}
	// fmt.Printf("\nService Basic Details [JSON]:\n%s\n", outputJson)

	return 
}

func GetClinicServiceBasicDetailsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")	

	//TODO: This line is an ad-hoc solution to remove the 
	//      std output triggered by CORS preflighted request via http OPTIONS.
	//      The correct solution is to add a reverse-proxy server to the
	//      main server so that all requests are channeled to this reverse proxy
	//      in order to overcome CORS restriction.
	if (r.Method == "OPTIONS") { return } 
	
	r.ParseForm()
	fmt.Println("[GetClinicServiceBasicDetailsHandler] Request Form Data Received! \n")
	fmt.Println(r.Form)
	
	userId := r.Form["userId"][0]
	userPwd := r.Form["userPwd"][0]
	deptName := r.Form["deptName"][0]
	clinicId := r.Form["clinicId"][0]
	district := r.Form["district"][0]
	state := r.Form["state"][0]
	
	clinicServiceBasicDetails, err := getClinicServiceBasicDetails(userId, userPwd, 
										deptName, clinicId, district, state)
	if err != nil {
		log.Print(err)
		w.WriteHeader(500)
		fmt.Fprintf(w, "Internal server error!")
		return
	}
	fmt.Printf("\nService Basic Details Response [JSON]:\n%s\n", clinicServiceBasicDetails)
	fmt.Fprintf(w, "%s", clinicServiceBasicDetails)
}

func getClinicServiceAdvDetails(isDirty bool, isEnabled bool, userId string, userPwd string, svcName string,
	deptName string, clinicId string, district string, state string) (outputJson []byte, err error) {
	
	l, err := ldap.DialURL("ldap://127.0.0.1:389")
	if err != nil {
		log.Print(err)
		return nil, err
	}
	defer l.Close()

	// TODO: KIV to change 'staffId' to 'userId' by changing
	//       the LDAP object class from 'staff' to regular 'user'
	userDN := fmt.Sprintf("staffId=%s,%s", userId, STAFF_BASE_DN)
	err = l.Bind(userDN, userPwd)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	serviceDN := fmt.Sprintf(SERVICE_TEMPLATE_DN3, svcName, deptName, clinicId, district, state)

	// If 'isDirty' is true, dept 'isEnabled' status
	// has been changed by client app.
	// Thus, we need to first update the dept 'isEnabled'
	// attribute value before retrieving dept advanced data,
	// and sending them to client app.
	if isDirty {
		modifyRequest := ldap.NewModifyRequest(
			serviceDN, 
			nil,
		)
		var isEnabledStr string
		if isEnabled {
			isEnabledStr = "1"
		} else {
			isEnabledStr = "0"
		}
		modifyRequest.Replace("clinicServiceIsEnabled", []string{isEnabledStr})

		err = l.Modify(modifyRequest)
		if err != nil {
			log.Print(err)
			return nil, err
		}
	} 

	// Search for the advanced data for the dept of the clinic
	searchFilter := "(&)"	
	searchRequest := ldap.NewSearchRequest(
		serviceDN,
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0,0, false,
		searchFilter,
		[]string{"*"},
		nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Print(err)
		resultCode := err.(*ldap.Error).ResultCode //uint16 
		if resultCode == 32 {
			//LDAPResultNoSuchObject
			fmt.Println("There is no such Service object found!")
		}	
		return nil, err
	}

	entry := sr.Entries[0]
	serviceAvaiDays := entry.GetAttributeValue("clinicServiceAvaiDays")
	serviceStartHrs := entry.GetAttributeValue("clinicServiceStartHour")
	serviceEndHrs := entry.GetAttributeValue("clinicServiceEndHour")
	serviceAvgConsultTime, _ := strconv.Atoi(entry.GetAttributeValue("clinicServiceAvgConsultTime"))

	serviceAdvDetails := ServiceAdvDetails{
		AvaiDays: serviceAvaiDays,
		StartHrs: serviceStartHrs,
		EndHrs: serviceEndHrs,
		AvgConsultTime: serviceAvgConsultTime,		
	}	

	outputJson, err = json.MarshalIndent(serviceAdvDetails, "", "\t")
	if err != nil {
		log.Print(err)
		return nil, err
	}
	// fmt.Printf("\nService Adv Details [JSON]:\n%s\n", outputJson)

	return 
}

func GetClinicServiceAdvDetailsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")	

	//TODO: This line is an ad-hoc solution to remove the 
	//      std output triggered by CORS preflighted request via http OPTIONS.
	//      The correct solution is to add a reverse-proxy server to the
	//      main server so that all requests are channeled to this reverse proxy
	//      in order to overcome CORS restriction.
	if (r.Method == "OPTIONS") { return } 
	
	r.ParseForm()
	fmt.Println("[GetClinicServiceAdvDetailsHandler] Request Form Data Received! \n")
	fmt.Println(r.Form)
	
	isDirty, _ := strconv.ParseBool(r.Form["isDirty"][0])
	isEnabled, _ := strconv.ParseBool(r.Form["isEnabled"][0])
	userId := r.Form["userId"][0]
	userPwd := r.Form["userPwd"][0]
	svcName := r.Form["svcName"][0]
	deptName := r.Form["deptName"][0]
	clinicId := r.Form["clinicId"][0]
	district := r.Form["district"][0]
	state := r.Form["state"][0]
	
	clinicServiceAdvDetails, err := getClinicServiceAdvDetails(isDirty, isEnabled, userId, userPwd, 
										svcName, deptName, clinicId, district, state)
	if err != nil {
		log.Print(err)
		w.WriteHeader(500)
		fmt.Fprintf(w, "Internal server error!")
		return
	}
	fmt.Printf("\nService Adv Details Response [JSON]:\n%s\n", clinicServiceAdvDetails)
	fmt.Fprintf(w, "%s", clinicServiceAdvDetails)
}

type UpsertDeptStruct struct {
	UserId string				`json:"userId"`
	UserPwd string 				`json:"userPwd"`
	State string 				`json:"state"`
	District string 			`json:"district"`
	ClinicName string			`json:"clinicName"`
	DeptName string				`json:"deptName"`
	DeptIsEnabled string		`json:"deptIsEnabled"`
	DeptAvaiDays string			`json:"deptAvaiDays"`
	DeptStartHrs string			`json:"deptStartHrs"`
	DeptEndHrs string			`json:"deptEndHrs"`
	DeptNumOfStaff string		`json:"deptNumOfStaff"`
	DeptStaffIds []string		`json:"deptStaffIds"`
	DeptMaxPt string 			`json:"deptMaxPt"`
}

func addDept(uds UpsertDeptStruct) (err error) {
    l, err := ldap.DialURL("ldap://127.0.0.1:389")
    if err != nil {
		log.Print(err)
		return
    }
    defer l.Close()

	// TODO: KIV to change 'staffId' to 'userId' by changing
	//       the LDAP object class from 'staff' to regular 'user'
	userDN := fmt.Sprintf("staffId=%s,%s", uds.UserId, STAFF_BASE_DN)
    err = l.Bind(userDN, uds.UserPwd)
    if err != nil {
		log.Print(err)
		return
    }

	newDeptDN := fmt.Sprintf(DEPT_TEMPLATE_DN, uds.DeptName, uds.ClinicName,
							uds.District, uds.State)
    addReq := ldap.NewAddRequest(
		newDeptDN, 
		nil,
	)
	addReq.Attribute("objectClass", []string{"top", "mytcaClinicDept"})
    addReq.Attribute("clinicDeptName", []string{uds.DeptName})
	addReq.Attribute("clinicDeptIsEnabled", []string{uds.DeptIsEnabled})
	addReq.Attribute("clinicDeptAvaiDays", []string{uds.DeptAvaiDays})
	addReq.Attribute("clinicDeptStartHour", []string{uds.DeptStartHrs})
	addReq.Attribute("clinicDeptEndHour", []string{uds.DeptEndHrs})
	addReq.Attribute("clinicDeptNumOfStaff", []string{uds.DeptNumOfStaff})
	addReq.Attribute("clinicDeptMaxPt", []string{uds.DeptMaxPt})
	if len(uds.DeptStaffIds) != 0 {
		addReq.Attribute("clinicDeptStaffId", uds.DeptStaffIds)
	}

    err = l.Add(addReq)
    if err != nil {
		log.Print(err)
		return
	}
	
	return nil
}

func addDeptHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")		
	r.ParseForm()

	//TODO: This line is an ad-hoc solution to remove the 
	//      std output triggered by CORS preflighted request via http OPTIONS.
	//      The correct solution is to add a reverse-proxy server to the
	//      main server so that all requests are channeled to this reverse proxy
	//      in order to overcome CORS restriction.
	if (r.Method == "OPTIONS") { return }

	fmt.Println("[addDeptHandler] Request Form Data Received!\n")
	fmt.Println(r.PostForm)

	for formData, _ := range r.PostForm {
		var uds UpsertDeptStruct
		err := json.Unmarshal([]byte(formData), &uds)
		if err != nil {
			log.Print(err)
			w.WriteHeader(500)
			fmt.Fprintf(w, "Internal server error!")
			return
		}
		fmt.Printf("UpsertDeptStruct: %+v \n", uds)

		err = addDept(uds)
		if err != nil {
			log.Print(err)
			w.WriteHeader(500)
			fmt.Fprintf(w, "Internal server error!")
			return
		}
		break
	}

	fmt.Println("Added new dept to OpenDJ directory.\nDone")	
}

func updateDept(uds UpsertDeptStruct) (err error) {
    l, err := ldap.DialURL("ldap://127.0.0.1:389")
    if err != nil {
		log.Print(err)
		return
    }
    defer l.Close()

	// TODO: KIV to change 'staffId' to 'userId' by changing
	//       the LDAP object class from 'staff' to regular 'user'
	userDN := fmt.Sprintf("staffId=%s,%s", uds.UserId, STAFF_BASE_DN)
    err = l.Bind(userDN, uds.UserPwd)
    if err != nil {
		log.Print(err)
		return
    }

	deptDN := fmt.Sprintf(DEPT_TEMPLATE_DN, uds.DeptName, uds.ClinicName,
							uds.District, uds.State)
    modifyReq := ldap.NewModifyRequest(
		deptDN, 
		nil,
	)
    modifyReq.Replace("clinicDeptName", []string{uds.DeptName})
	modifyReq.Replace("clinicDeptAvaiDays", []string{uds.DeptAvaiDays})
	modifyReq.Replace("clinicDeptStartHour", []string{uds.DeptStartHrs})
	modifyReq.Replace("clinicDeptEndHour", []string{uds.DeptEndHrs})
	modifyReq.Replace("clinicDeptNumOfStaff", []string{uds.DeptNumOfStaff})
	modifyReq.Replace("clinicDeptIsEnabled", []string{uds.DeptIsEnabled})
	modifyReq.Replace("clinicDeptMaxPt", []string{uds.DeptMaxPt})
	if len(uds.DeptStaffIds) != 0 {
		modifyReq.Replace("clinicDeptStaffId", uds.DeptStaffIds)
	} else {
		modifyReq.Replace("clinicDeptStaffId", []string{})
	}
	
    err = l.Modify(modifyReq)
    if err != nil {
		log.Print(err)
		return
	}
	
	return nil
}

func updateDeptHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")		
	r.ParseForm()

	//TODO: This line is an ad-hoc solution to remove the 
	//      std output triggered by CORS preflighted request via http OPTIONS.
	//      The correct solution is to add a reverse-proxy server to the
	//      main server so that all requests are channeled to this reverse proxy
	//      in order to overcome CORS restriction.
	if (r.Method == "OPTIONS") { return }

	fmt.Println("[updateDeptHandler] Request Form Data Received!\n")
	fmt.Println(r.PostForm)

	for formData, _ := range r.PostForm {
		var uds UpsertDeptStruct
		err := json.Unmarshal([]byte(formData), &uds)
		if err != nil {
			log.Print(err)
			w.WriteHeader(500)
			fmt.Fprintf(w, "Internal server error!")
			return
		}
		fmt.Printf("UpsertDeptStruct: %+v \n", uds)

		err = updateDept(uds)
		if err != nil {
			log.Print(err)
			w.WriteHeader(500)
			fmt.Fprintf(w, "Internal server error!")
			return
		}
		break
	}

	fmt.Println("Updated dept in OpenDJ directory.\nDone")	
}

type UpsertSvcStruct struct {
	UserId string				`json:"userId"`
	UserPwd string 				`json:"userPwd"`
	State string 				`json:"state"`
	District string 			`json:"district"`
	ClinicId string				`json:"clinicId"`
	DeptName string				`json:"deptName"`
	SvcName string 				`json:"svcName"`
	SvcIsEnabled string			`json:"svcIsEnabled"`
	SvcAvaiDays string			`json:"svcAvaiDays"`
	SvcStartHrs string			`json:"svcStartHrs"`
	SvcEndHrs string			`json:"svcEndHrs"`
	SvcAvgConsultTime int 		`json:"svcAvgConsultTime"`
}

func addSvc(uss UpsertSvcStruct) (err error) {
    l, err := ldap.DialURL("ldap://127.0.0.1:389")
    if err != nil {
		log.Print(err)
		return
    }
    defer l.Close()

	// TODO: KIV to change 'staffId' to 'userId' by changing
	//       the LDAP object class from 'staff' to regular 'user'
	userDN := fmt.Sprintf("staffId=%s,%s", uss.UserId, STAFF_BASE_DN)
    err = l.Bind(userDN, uss.UserPwd)
    if err != nil {
		log.Print(err)
		return
    }

	newSvcDN := fmt.Sprintf(SERVICE_TEMPLATE_DN3, uss.SvcName, uss.DeptName, uss.ClinicId,
							uss.District, uss.State)
    addReq := ldap.NewAddRequest(
		newSvcDN, 
		nil,
	)
	addReq.Attribute("objectClass", []string{"top", "mytcaClinicService"})
    addReq.Attribute("clinicServiceName", []string{uss.SvcName})
	addReq.Attribute("clinicServiceIsEnabled", []string{uss.SvcIsEnabled})
	addReq.Attribute("clinicServiceAvaiDays", []string{uss.SvcAvaiDays})
	addReq.Attribute("clinicServiceStartHour", []string{uss.SvcStartHrs})
	addReq.Attribute("clinicServiceEndHour", []string{uss.SvcEndHrs})

    err = l.Add(addReq)
    if err != nil {
		log.Print(err)
		return
	}
	
	return nil
}

func addSvcHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")		
	r.ParseForm()

	//TODO: This line is an ad-hoc solution to remove the 
	//      std output triggered by CORS preflighted request via http OPTIONS.
	//      The correct solution is to add a reverse-proxy server to the
	//      main server so that all requests are channeled to this reverse proxy
	//      in order to overcome CORS restriction.
	if (r.Method == "OPTIONS") { return }

	fmt.Println("[addSvcHandler] Request Form Data Received!\n")
	fmt.Println(r.PostForm)

	for formData, _ := range r.PostForm {
		var uss UpsertSvcStruct
		err := json.Unmarshal([]byte(formData), &uss)
		if err != nil {
			log.Print(err)
			w.WriteHeader(500)
			fmt.Fprintf(w, "Internal server error!")
			return
		}
		fmt.Printf("UpsertSvcStruct: %+v \n", uss)

		err = addSvc(uss)
		if err != nil {
			log.Print(err)
			w.WriteHeader(500)
			fmt.Fprintf(w, "Internal server error!")
			return
		}
		fmt.Printf("Added new %s service to %s dept in OpenDJ directory.\nDone \n",
					uss.SvcName, uss.DeptName)	
		break
	}	
}

func updateSvc(uss UpsertSvcStruct) (err error) {
    l, err := ldap.DialURL("ldap://127.0.0.1:389")
    if err != nil {
		log.Print(err)
		return
    }
    defer l.Close()

	// TODO: KIV to change 'staffId' to 'userId' by changing
	//       the LDAP object class from 'staff' to regular 'user'
	userDN := fmt.Sprintf("staffId=%s,%s", uss.UserId, STAFF_BASE_DN)
    err = l.Bind(userDN, uss.UserPwd)
    if err != nil {
		log.Print(err)
		return
    }

	svcDN := fmt.Sprintf(SERVICE_TEMPLATE_DN3, uss.SvcName, uss.DeptName, uss.ClinicId,
							uss.District, uss.State)
    modifyReq := ldap.NewModifyRequest(
		svcDN, 
		nil,
	)
	modifyReq.Replace("clinicServiceIsEnabled", []string{uss.SvcIsEnabled})
	modifyReq.Replace("clinicServiceAvaiDays", []string{uss.SvcAvaiDays})
	modifyReq.Replace("clinicServiceStartHour", []string{uss.SvcStartHrs})
	modifyReq.Replace("clinicServiceEndHour", []string{uss.SvcEndHrs})	
	
    err = l.Modify(modifyReq)
    if err != nil {
		log.Print(err)
		return
	}
	
	return nil
}

func updateSvcHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")		
	r.ParseForm()

	//TODO: This line is an ad-hoc solution to remove the 
	//      std output triggered by CORS preflighted request via http OPTIONS.
	//      The correct solution is to add a reverse-proxy server to the
	//      main server so that all requests are channeled to this reverse proxy
	//      in order to overcome CORS restriction.
	if (r.Method == "OPTIONS") { return }

	fmt.Println("[updateSvcHandler] Request Form Data Received!\n")
	fmt.Println(r.PostForm)

	for formData, _ := range r.PostForm {
		var uss UpsertSvcStruct
		err := json.Unmarshal([]byte(formData), &uss)
		if err != nil {
			log.Print(err)
			w.WriteHeader(500)
			fmt.Fprintf(w, "Internal server error!")
			return
		}
		fmt.Printf("UpsertSvcStruct: %+v \n", uss)

		err = updateSvc(uss)
		if err != nil {
			log.Print(err)
			w.WriteHeader(500)
			fmt.Fprintf(w, "Internal server error!")
			return
		}
		fmt.Printf("Updated %s service to %s dept in OpenDJ directory.\nDone \n",
					uss.SvcName, uss.DeptName)
		break
	}
}

type UpsertClinicStruct struct {
	UserId string				`json:"userId"`
	UserPwd string 				`json:"userPwd"`
	State string 				`json:"state"`
	District string 			`json:"district"`
	ClinicId string				`json:"clinicId"`
	ClinicName string			`json:"clinicName"`
	CloseDays string			`json:"closeDays"`
	PublicHolidays []string		`json:"publicHolidays"`
	StaffIds []string			`json:"staffIds"`
}

func updateClinic(ucs UpsertClinicStruct) (err error) {
    l, err := ldap.DialURL("ldap://127.0.0.1:389")
    if err != nil {
        log.Print(err)
		return
    }
	defer l.Close()
	
	// TODO: KIV to change 'staffId' to 'userId' by changing
	//       the LDAP object class from 'staff' to regular 'user'
	userDN := fmt.Sprintf("staffId=%s,%s", ucs.UserId, STAFF_BASE_DN)
	err = l.Bind(userDN, ucs.UserPwd)
	if err != nil {
		log.Print(err)
		return
	}
	
	clinicDN := fmt.Sprintf(CLINIC_TEMPLATE_DN, ucs.ClinicId, 
							ucs.District, ucs.State)
	modifyReq := ldap.NewModifyRequest(
		clinicDN, 
		nil,
	)
	modifyReq.Replace("clinicCloseDays", []string{ucs.CloseDays})
	modifyReq.Replace("publicHolByMonth", ucs.PublicHolidays)
	modifyReq.Replace("staffId", ucs.StaffIds)

	err = l.Modify(modifyReq)
    if err != nil {
		log.Print(err)
		return
	}
	
	return nil
}

func updateClinicHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")		
	r.ParseForm()

	//TODO: This line is an ad-hoc solution to remove the 
	//      std output triggered by CORS preflighted request via http OPTIONS.
	//      The correct solution is to add a reverse-proxy server to the
	//      main server so that all requests are channeled to this reverse proxy
	//      in order to overcome CORS restriction.
	if (r.Method == "OPTIONS") { return }

	fmt.Println("[updateClinicHandler] Request Form Data Received!\n")
	fmt.Println(r.PostForm)

	for formData, _ := range r.PostForm {
		var ucs UpsertClinicStruct
		err := json.Unmarshal([]byte(formData), &ucs)
		if err != nil {
			log.Print(err)
			w.WriteHeader(500)
			fmt.Fprintf(w, "Internal server error!")
			return
		}
		fmt.Printf("UpsertClinicStruct: %+v \n", ucs)

		err = updateClinic(ucs)
		if err != nil {
			log.Print(err)
			w.WriteHeader(500)
			fmt.Fprintf(w, "Internal server error!")
			return
		}
		fmt.Printf("Updated %s ClinicId in OpenDJ directory.\nDone \n",
					ucs.ClinicId)
		break
	}
}