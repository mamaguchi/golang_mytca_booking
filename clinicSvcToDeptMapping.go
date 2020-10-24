package main

const (
	// Department
	DEPT_OPD_DISEASE			=	"opd_disease"
	DEPT_METHADONE_PHARMACIST	=	"methadone_pharmacist"
	DEPT_BLOOD_TAKING			=	"blood_taking"
	DEPT_WOUND_CARE				=	"wound_care"
	DEPT_FUNDOSCOPY				=	"fundoscopy"
	DEPT_XRAY					=	"x-ray"
	DEPT_DIET_COUNSELLING		=	"diet_counselling"
	DEPT_PREMARITAL_SCREENING	=	"premarital_screening"
	DEPT_MEDICAL_CHECKUP		=	"medical_checkup"
	DEPT_ANTENATAL				=	"antenatal"
	DEPT_POSTNATAL 				=	"postnatal"
	DEPT_FAMILY_PLANNING		=	"family_planning"
	DEPT_PAP_SMEAR				=	"pap_smear"
	DEPT_IMMUNIZATION			=	"immunization"
	
	// Service
	SVC_FEVER					=	"fever"
	SVC_URTI					=	"urti"
	SVC_ASTHMA					=	"asthma"
	SVC_TUBERCULOSIS			=	"tuberculosis"
	SVC_ABD_DISCOMFORT			=	"abdDiscomfort"
	SVC_AGE						=	"age"
	SVC_DIABETES				=	"diabetes"
	SVC_HYPERTENSION			=	"hypertension"
	SVC_PSY 					=	"psy"
	SVC_KBM						=	"kbm"
	SVC_METHADONE				=	"methadone"
	SVC_BLOOD_TAKING			=	"bloodTaking"
	SVC_WOUND_CARE				=	"woundCare"
	SVC_FUNDOSCOPY				=	"fundoscopy"
	SVC_XRAY					=	"x-ray"
	SVC_DIET_COUNSELLING		=	"dietCounselling"
	SVC_PREMARITAL_SCREENING	=	"premaritalScreening"
	SVC_MEDICAL_CHECKUP			=	"medicalCheckup"
	SVC_ANTENATAL_BOOKING		=	"antenatalBooking"
	SVC_ANTENATAL_FUP			=	"antenatalFUp"
	SVC_POSTNATAL				=	"postnatal"
	SVC_FAMILY_PLANNING			= 	"familyPlanning"
	SVC_PAP_SMEAR				=	"papSmear"
	SVC_IMMUNIZATION			=	"immunization"
	SVC_OTHER_PROBS				=	"otherProblems"
)

var SvcToDeptMap map[string]string

func init() {
	SvcToDeptMap = make(map[string]string)
	
	// Create Service-to-Department Mapping
	SvcToDeptMap[SVC_FEVER]					= DEPT_OPD_DISEASE
	SvcToDeptMap[SVC_URTI]					= DEPT_OPD_DISEASE
	SvcToDeptMap[SVC_ASTHMA]				= DEPT_OPD_DISEASE
	SvcToDeptMap[SVC_TUBERCULOSIS]			= DEPT_OPD_DISEASE
	SvcToDeptMap[SVC_ABD_DISCOMFORT]		= DEPT_OPD_DISEASE
	SvcToDeptMap[SVC_AGE]					= DEPT_OPD_DISEASE
	SvcToDeptMap[SVC_DIABETES] 				= DEPT_OPD_DISEASE
	SvcToDeptMap[SVC_HYPERTENSION]  		= DEPT_OPD_DISEASE
	SvcToDeptMap[SVC_PSY]					= DEPT_OPD_DISEASE
	SvcToDeptMap[SVC_KBM]					= DEPT_OPD_DISEASE
	SvcToDeptMap[SVC_METHADONE]				= DEPT_METHADONE_PHARMACIST 
	SvcToDeptMap[SVC_BLOOD_TAKING]			= DEPT_BLOOD_TAKING
	SvcToDeptMap[SVC_WOUND_CARE]			= DEPT_WOUND_CARE
	SvcToDeptMap[SVC_FUNDOSCOPY]   			= DEPT_FUNDOSCOPY
	SvcToDeptMap[SVC_XRAY]    				= DEPT_XRAY
	SvcToDeptMap[SVC_DIET_COUNSELLING]		= DEPT_DIET_COUNSELLING
	SvcToDeptMap[SVC_PREMARITAL_SCREENING]	= DEPT_PREMARITAL_SCREENING
	SvcToDeptMap[SVC_MEDICAL_CHECKUP]		= DEPT_MEDICAL_CHECKUP
	SvcToDeptMap[SVC_ANTENATAL_BOOKING]		= DEPT_ANTENATAL
	SvcToDeptMap[SVC_ANTENATAL_FUP]			= DEPT_ANTENATAL
	SvcToDeptMap[SVC_POSTNATAL]				= DEPT_POSTNATAL
	SvcToDeptMap[SVC_FAMILY_PLANNING]		= DEPT_FAMILY_PLANNING
	SvcToDeptMap[SVC_PAP_SMEAR]				= DEPT_PAP_SMEAR
	SvcToDeptMap[SVC_IMMUNIZATION]			= DEPT_IMMUNIZATION
	SvcToDeptMap[SVC_OTHER_PROBS]			= DEPT_OPD_DISEASE
}
