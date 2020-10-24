package main

import (
	"log"
	"fmt"
	"strconv"
	"strings"
	"encoding/json"
	"github.com/go-ldap/ldap/v3"
)

const (
	DIR_MGR_DN      	 	=	"cn=Directory Manager"
	DIR_MGR_PWD    		 	=	"88motherfaker88"
	STAFF_BASE_DN    	 	=	"ou=people,dc=example,dc=com"
	CLINIC_BASE_DN   	 	=   "ou=kkm-clinic,ou=groups,dc=example,dc=com"
	CLINIC_TEMPLATE_DN   	=   "cn=%s,ou=pkd_%s,ou=jkn_%s," + CLINIC_BASE_DN
	DEPT_BASE_TEMPLATE_DN   =   "ou=dept," + CLINIC_TEMPLATE_DN
	DEPT_TEMPLATE_DN   		=   "clinicDeptName=%s," + DEPT_BASE_TEMPLATE_DN
	SERVICE_BASE_TEMPLATE_DN =   "ou=service," + DEPT_TEMPLATE_DN
	SERVICE_TEMPLATE_DN  	=   "clinicServiceName=%s,ou=service," + CLINIC_TEMPLATE_DN
	SERVICE_TEMPLATE_DN2 	=   "clinicServiceName=%s,ou=service,clinicDeptName=%s,ou=dept," + CLINIC_TEMPLATE_DN
	SERVICE_TEMPLATE_DN3 	=   "clinicServiceName=%s,ou=service," + DEPT_TEMPLATE_DN
)

type ClinicsDirectory struct {
	States []State 			`json:"states"`
}

type State struct {
	Name string				`json:"name"`
	Districts []District 	`json:"districts"`
}

type District struct {
	Name string				`json:"name"`
	Clinics []Clinic 		`json:"clinics"`
}

type Clinic struct {
	Name string 			`json:"name"`
	Id string				`json:"id"`
}

type ClinicServiceMeta struct {
	NumOfStaff int 				`json:"numOfStaff"`
	AvaiDays string				`json:"avaiDays"'`
	StartOpHrs string 			`json:"startOpHrs"`
	EndOpHrs string				`json:"endOpHrs"`
}

// States
var perlis = State{Name: "Perlis"}
var kedah = State{Name: "Kedah"}
var pulauPinang = State{Name: "Pulau Pinang"}
var perak = State{Name: "Perak"}
var selangor = State{Name: "Selangor"}
var negeriSembilan = State{Name: "Negeri Sembilan"}
var melaka = State{Name: "Melaka"}
var johor = State{Name: "Johor"}
var pahang = State{Name: "Pahang"}
var kelantan = State{Name: "Kelantan"}
var terengganu = State{Name: "Terengganu"}
var sabah = State{Name: "Sabah"}
var sarawak = State{Name: "Sarawak"}
var pulauLabuan = State{Name: "Pulau Labuan"}

// Districts
var maran = District{Name: "Maran"}
var termeloh = District{Name: "Termeloh"}
var klang = District{Name: "Klang"}

// Clinic Directory
var ClinicsDir ClinicsDirectory

// Init
func init() {
	pahang.Districts = []District{maran, termeloh}
	selangor.Districts = []District{klang}

	states := []State{perlis, kedah, pulauPinang, perak, selangor,
						negeriSembilan, melaka, johor, pahang, kelantan,
						terengganu, sabah, sarawak, pulauLabuan}

	ClinicsDir.States = states
}


// Non-exported Func
// =================

// Get a list of clinics name and id via OpenDJ LDAP.
func getClinicsNameAndId() (clinicsDirCpy ClinicsDirectory, err error) {
	l, err := ldap.DialURL("ldap://127.0.0.1:389")
    if err != nil {
            log.Fatal(err)
    }
    defer l.Close()

	err = l.Bind(DIR_MGR_DN, DIR_MGR_PWD)
    if err != nil {
		return 
	}

	clinicsDirCpy = ClinicsDir

	for i, state := range clinicsDirCpy.States {
		stateName := state.Name

		for j, district := range state.Districts {
			districtName := district.Name

			clinicSearchDN := fmt.Sprintf("ou=pkd_%s,ou=jkn_%s,%s",
										districtName, stateName, CLINIC_BASE_DN)
			searchRequest := ldap.NewSearchRequest(
				clinicSearchDN,
				ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0,0, false,
				"(&)",
				[]string{"clinicName", "cn"},
				nil,
			)
			sr, err := l.Search(searchRequest)                     
			if err != nil {
					log.Fatal(err)
			}
			for _, entry := range sr.Entries {
				clinicName := entry.GetAttributeValue("clinicName")
				clinicId := entry.GetAttributeValue("cn")
				clinic := Clinic{Name: clinicName, Id: clinicId}

				clinicsDirCpy.States[i].Districts[j].Clinics = append(clinicsDirCpy.States[i].Districts[j].Clinics, clinic)
			}
		}
	}

	return 
}


// Exported Func
// =============

// Create a directory of all the clinics in every district
// from every states in Malaysia.
func GetClinicsDirectory() (output []byte, err error) {
	clinicsDirCpy, err := getClinicsNameAndId()
	if err != nil {
		return nil, err
	}
	output, err = json.MarshalIndent(clinicsDirCpy, "", "\t")
	return 
}

// Get the clinic service metadata via OpenDJ LDAP.
func GetClinicServiceMeta(service string, clinic string, district string, state string) (serviceMeta ClinicServiceMeta, err error) {
	l, err := ldap.DialURL("ldap://127.0.0.1:389")
    if err != nil {
            log.Fatal(err)
    }
    defer l.Close()

	err = l.Bind(DIR_MGR_DN, DIR_MGR_PWD)
    if err != nil {
		return 
	}

	serviceSearchDN := fmt.Sprintf(SERVICE_TEMPLATE_DN,
							service, clinic, district, state)
	searchRequest := ldap.NewSearchRequest(
		serviceSearchDN,
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0,0, false,
		"(&)",
		[]string{"clinicServiceNumOfStaff",
				"clinicServiceAvaiDays", 
				"clinicServiceStartHour",
				"clinicServiceEndHour"},
		nil,
	)
	sr, err := l.Search(searchRequest)                     
	if err != nil {
			log.Fatal(err)
	}

	for _, entry := range sr.Entries {
		serviceMeta.NumOfStaff, _ = strconv.Atoi(entry.GetAttributeValue("clinicServiceNumOfStaff"))
		serviceMeta.AvaiDays = entry.GetAttributeValue("clinicServiceAvaiDays")
		serviceMeta.StartOpHrs = entry.GetAttributeValue("clinicServiceStartHour")
		serviceMeta.EndOpHrs = entry.GetAttributeValue("clinicServiceEndHour")
	}

	return serviceMeta, nil
}

type ClinicDeptAndSvcData struct {
	Type string
	//
	DeptName string
	DeptNumOfStaff int
	DeptStartHrs string
	DeptEndHrs string 
	// DeptMaxPt int
	//
	SvcName string
	SvcNumOfStaff int
	SvcAvaiDays string
	// SvcStartHrs string
	// SvcEndHrs string
}

func GetClinicDeptAndServicesMeta(clinic string, district string, state string) (map[string][]ClinicDeptAndSvcData, error) {
	l, err := ldap.DialURL("ldap://127.0.0.1:389")
    if err != nil {
            log.Fatal(err)
    }
    defer l.Close()

	err = l.Bind(DIR_MGR_DN, DIR_MGR_PWD)
    if err != nil {
		return nil, err
	}

	searchDN := fmt.Sprintf(DEPT_BASE_TEMPLATE_DN,
							clinic, district, state)
	searchFilter := "(|(clinicDeptName=*)(clinicServiceName=*))"							
	searchRequest := ldap.NewSearchRequest(
		searchDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0,0, false,
		searchFilter,
		[]string{"*"}, //TODO: Make the projection more specific to the required attributes.
		nil,
	)
	sr, err := l.Search(searchRequest)                     
	if err != nil {
			log.Fatal(err)
	}

	m := make(map[string][]ClinicDeptAndSvcData)
	for _, entry := range sr.Entries {
		 if entry.GetAttributeValue("clinicDeptName") != "" {
			clinicDeptDN := entry.DN
			clinicDeptName := entry.GetAttributeValue("clinicDeptName")
			clinicDeptNumOfStaff, _ := strconv.Atoi(entry.GetAttributeValue("clinicDeptNumOfStaff"))
			clinicDeptStartHrs := entry.GetAttributeValue("clinicDeptStartHour")
			clinicDeptEndHrs := entry.GetAttributeValue("clinicDeptEndHour")
			clinicDeptMaxPt, _ := strconv.Atoi(entry.GetAttributeValue("clinicDeptMaxPt"))
			fmt.Println("Clinic Dept: \n", clinicDeptDN, clinicDeptName, clinicDeptNumOfStaff,
										clinicDeptStartHrs, clinicDeptEndHrs, clinicDeptMaxPt)
			
			meta := ClinicDeptAndSvcData{
				Type: "dept",
				DeptName: clinicDeptName,
				DeptNumOfStaff: clinicDeptNumOfStaff,
				DeptStartHrs: clinicDeptStartHrs,
				DeptEndHrs: clinicDeptEndHrs,
				// DeptMaxPt: clinicDeptMaxPt,
			}
			commonDN := clinicDeptDN
			m[commonDN] = append(m[commonDN], meta)

		} else if entry.GetAttributeValue("clinicServiceName") != "" {
			clinicSvcDN := entry.DN
			clinicSvcName := entry.GetAttributeValue("clinicServiceName")
			clinicSvcAvaiDays := entry.GetAttributeValue("clinicServiceAvaiDays")   
			// clinicSvcNumOfStaff, _ := strconv.Atoi(entry.GetAttributeValue("clinicServiceNumOfStaff"))
			// clinicSvcStartHrs := entry.GetAttributeValue("clinicServiceStartHour")  //Probably not needed here for now
			// clinicSvcEndHrs := entry.GetAttributeValue("clinicServiceEndHour")      //Probably not needed here for now
			fmt.Println("Clinic Service: \n", clinicSvcDN, clinicSvcName, clinicSvcAvaiDays)
									
			meta := ClinicDeptAndSvcData{
				Type: "svc",
				SvcName: clinicSvcName,
				SvcAvaiDays: clinicSvcAvaiDays,
				// SvcNumOfStaff: clinicSvcNumOfStaff,
				// SvcStartHrs: clinicSvcStartHrs,
				// SvcEndHrs: clinicSvcEndHrs,
			}
			sepIdx := strings.Index(clinicSvcDN, ",")
			shortenClinicSvcDN := clinicSvcDN[(sepIdx+1):]
			sepIdx = strings.Index(shortenClinicSvcDN, ",")
			commonDN := shortenClinicSvcDN[(sepIdx+1):]
			m[commonDN] = append(m[commonDN], meta)
		}
	}

	return m, nil
}

func checkIfSvcExist(svc string, dept string, clinic string, district string, state string) (bool, error) {
	l, err := ldap.DialURL("ldap://127.0.0.1:389")
    if err != nil {
            log.Fatal(err)
    }
    defer l.Close()

	err = l.Bind(DIR_MGR_DN, DIR_MGR_PWD)
    if err != nil {
		return false, err
	}

	searchDN := fmt.Sprintf(SERVICE_TEMPLATE_DN2,
							svc, dept, clinic, district, state)
	searchRequest := ldap.NewSearchRequest(
		searchDN,
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0,0, false,
		"(&)",
		[]string{"clinicServiceName"}, 
		nil,
	)
	sr, err := l.Search(searchRequest)                     
	if err != nil {
		// If the entry is not found in LDAP directory, the 
		// search request will return the following error:
		// """
		// LDAP Result Code 32 "No Such Object"
		// """
		// So we handle the error here accordingly and 
		// return false for search result.
		return false, nil
	}

	for _, entry := range sr.Entries {
		if entry.GetAttributeValue("clinicServiceName") == svc {
			return true, nil 
		}
	}
	return false, nil
}

type ClinicSvcOpHrsMeta struct {
	DeptOpHrs DeptOpHrsMeta		`json:"deptOpHrsMeta"`
	SvcOpHrs SvcOpHrsMeta 		`json:"svcOpHrsMeta"`
}

type DeptOpHrsMeta struct {
	StartHrs string				`json:"startHrs"` 
	EndHrs string				`json:"endHrs"`
}

type SvcOpHrsMeta struct {
	StartHrs string				`json:"startHrs"` 
	EndHrs string				`json:"endHrs"`
}

func GetClinicSvcOpHrs(service string, dept string, clinic string, district string, 
						state string) (clinicSvcOpHrsMeta ClinicSvcOpHrsMeta, err error) {
	l, err := ldap.DialURL("ldap://127.0.0.1:389")
    if err != nil {
            log.Fatal(err)
    }
    defer l.Close()

	err = l.Bind(DIR_MGR_DN, DIR_MGR_PWD)
    if err != nil {
		return 
	}

	searchDN := fmt.Sprintf(DEPT_TEMPLATE_DN, dept, clinic, district, state)
	searchFilter := fmt.Sprintf("(|(clinicDeptName=%s)(clinicServiceName=%s))", dept, service)
	searchRequest := ldap.NewSearchRequest(
		searchDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0,0, false,
		searchFilter,
		[]string{"clinicDeptName", "clinicDeptStartHour", "clinicDeptEndHour", 
				 "clinicServiceName", "clinicServiceStartHour", "clinicServiceEndHour"}, 
		nil,
	)

	sr, err := l.Search(searchRequest)                     
	if err != nil {
			log.Fatal(err)
	}

	for _, entry := range sr.Entries {
		if entry.GetAttributeValue("clinicDeptName") != "" {
			clinicSvcOpHrsMeta.DeptOpHrs = DeptOpHrsMeta{
				StartHrs: entry.GetAttributeValue("clinicDeptStartHour"),
				EndHrs: entry.GetAttributeValue("clinicDeptEndHour"),
			}	
		} else if entry.GetAttributeValue("clinicServiceName") != "" {
			clinicSvcOpHrsMeta.SvcOpHrs = SvcOpHrsMeta{
				StartHrs: entry.GetAttributeValue("clinicServiceStartHour"),
				EndHrs: entry.GetAttributeValue("clinicServiceEndHour"),
			}
		}
	}
	return clinicSvcOpHrsMeta, nil
}

