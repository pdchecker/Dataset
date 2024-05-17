package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"log"
	"time"
)

type AppointmentContract struct {
	contractapi.Contract
	AppointmentCount int
}

type DoctorAvailability struct {
	DoctorID       string              `json:"doctorID"`
	Availability   map[string][]string `json:"availability"`
	TimeSlotLength int                 `json:"timeSlotLength"`
}

type Appointment struct {
	AppointmentID   string `json:"appointmentID"`
	PatientID       string `json:"patientID"`
	DoctorID        string `json:"doctorID"`
	AppointmentDate string `json:"appointmentDate"`
	AppointmentTime string `json:"AppointmentTime"`
	Status          string `json:"status"`
}

type Patient struct {
	PatientID    string `json:"patientID"`
	Name         string `json:"name"`
	Age          int    `json:"age"`
	Gender       string `json:"gender"`
	ICNo         string `json:"ic_no"`
	MobileNumber string `json:"mobile_number"`
}

// InitPatients initiates a few patients information
func (ac *AppointmentContract) InitPatients(ctx contractapi.TransactionContextInterface) error {
	// Create some patient information
	patients := []Patient{
		{PatientID: "patient1", Name: "John Doe", Age: 30, Gender: "Male", ICNo: "123456789", MobileNumber: "1234567890"},
		{PatientID: "patient2", Name: "Jane Smith", Age: 35, Gender: "Female", ICNo: "987654321", MobileNumber: "9876543210"},
		{PatientID: "patient3", Name: "David Johnson", Age: 40, Gender: "Male", ICNo: "543216789", MobileNumber: "5432167890"},
		{PatientID: "patient4", Name: "Emily Brown", Age: 25, Gender: "Female", ICNo: "789456123", MobileNumber: "7894561230"},
		{PatientID: "patient5", Name: "Michael Wilson", Age: 45, Gender: "Male", ICNo: "654789321", MobileNumber: "6547893210"},
		{PatientID: "patient6", Name: "Sarah Taylor", Age: 50, Gender: "Female", ICNo: "456123789", MobileNumber: "4561237890"},
	}

	// Store patient information on the blockchain
	for _, patient := range patients {
		patientJSON, err := json.Marshal(patient)
		if err != nil {
			return fmt.Errorf("failed to marshal patient JSON: %v", err)
		}

		err = ctx.GetStub().PutState(patient.PatientID, patientJSON)
		if err != nil {
			return fmt.Errorf("failed to put patient to ledger: %v", err)
		}
	}

	return nil
}

// SetDoctorAvailability set doctor availability time
func (ac *AppointmentContract) SetDoctorAvailability(ctx contractapi.TransactionContextInterface, doctorID string) error {

	layout := "2006-01-02"

	availability := make(map[string][]string)
	//Set the length of the appointment
	timeSlotLength := 20

	//get current date
	now := time.Now()
	// get Monday current of current week
	monday := now.AddDate(0, 0, -int(now.Weekday())+1)

	//Calculate next two weeks availability
	for day := 0; day < 14; day++ {
		//calculate date
		date := monday.AddDate(0, 0, day)
		//Skip Saturdays and Sundays
		if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
			continue
		}
		//Generate time slot and store
		availability[date.Format(layout)] = generateTimeSlots()
	}

	doctorAvailability := &DoctorAvailability{
		DoctorID:       doctorID,
		Availability:   availability,
		TimeSlotLength: timeSlotLength,
	}

	doctorAvailabilityJSON, err := json.Marshal(doctorAvailability)
	if err != nil {
		return fmt.Errorf("failed to marshal doctor availability: %v", err)
	}
	err = ctx.GetStub().PutState(doctorID, doctorAvailabilityJSON)
	if err != nil {
		return fmt.Errorf("failed to put doctor availability to ledger: %v", err)
	}

	return nil
}

// generateTimeSlots generate time slot
func generateTimeSlots() []string {
	timeSlots := make([]string, 0)
	slotDuration := 20 * time.Minute // The length of the time slot is 20 minutes

	startTimeMorning := time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)
	endTimeMorning := time.Date(0, 1, 1, 12, 0, 0, 0, time.UTC)

	for startTimeMorning.Before(endTimeMorning) {
		timeSlots = append(timeSlots, startTimeMorning.Format("15:04"))
		startTimeMorning = startTimeMorning.Add(slotDuration)
	}

	startTimeAfternoon := time.Date(0, 1, 1, 14, 0, 0, 0, time.UTC)
	endTimeAfternoon := time.Date(0, 1, 1, 18, 0, 0, 0, time.UTC)
	for startTimeAfternoon.Before(endTimeAfternoon) {
		timeSlots = append(timeSlots, startTimeAfternoon.Format("15:04"))
		startTimeAfternoon = startTimeAfternoon.Add(slotDuration)
	}

	return timeSlots
}

// GetDoctorAvailability retrieve doctor's available
func (ac *AppointmentContract) GetDoctorAvailability(ctx contractapi.TransactionContextInterface, doctorID string) (map[string][]string, error) {

	doctorAvailabilityJSON, err := ctx.GetStub().GetState(doctorID)

	if err != nil {
		return nil, fmt.Errorf("failed to read doctor availability from ledger: %v", err)
	}
	if doctorAvailabilityJSON == nil {
		return nil, fmt.Errorf("doctor availability not found for doctor ID %s", doctorID)
	}

	var doctorAvailability DoctorAvailability
	err = json.Unmarshal(doctorAvailabilityJSON, &doctorAvailability)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal doctor availability: %v", err)
	}

	return doctorAvailability.Availability, nil
	// Example: "2024-02-19": {"09:00", "09:20", "09:40", "10:00", "10:20", "10:40", "14:00", "14:20", "14:40", "15:00", "15:20", "15:40"}
}

// BookAppointment books an appointment for a patient with a doctor
func (ac *AppointmentContract) BookAppointment(ctx contractapi.TransactionContextInterface, doctorID string,
	patientID string, appointmentDate string, appointmentTime string) error {

	//Getting Doctor Availability
	availability, err := ac.GetDoctorAvailability(ctx, doctorID)
	if err != nil {
		return fmt.Errorf("failed to get doctor availability: %v", err)
	}

	//Check the doctor's availability for that appointment date
	timeSlots, ok := availability[appointmentDate]
	if !ok {
		return fmt.Errorf("doctor is not available on %s", appointmentDate)
	}

	//Check availability of appointment time
	var available bool
	for _, slot := range timeSlots {
		if slot == appointmentTime {
			available = true
			break
		}
	}
	if !available {
		return fmt.Errorf("doctor is not available at %s on %s", appointmentTime, appointmentDate)
	}

	// Remove appointment slot from availability time slot
	var matchIndex int = -1
	for i, slot := range timeSlots {
		if slot == appointmentTime {
			matchIndex = i
			break
		}
	}
	if matchIndex != -1 {
		availability[appointmentDate] = append(timeSlots[:matchIndex], timeSlots[matchIndex+1:]...)
	}

	//update doctor availability to ledger
	err = ac.UpdateDoctorAvailability(ctx, doctorID, availability)
	if err != nil {
		return fmt.Errorf("failed to update doctor availability: %v", err)
	}

	//Create appointment object
	appointmentID := ac.generateAppointmentID()

	appointment := &Appointment{
		AppointmentID:   appointmentID,
		PatientID:       patientID,
		DoctorID:        doctorID,
		AppointmentDate: appointmentDate,
		AppointmentTime: appointmentTime,
		Status:          "Scheduled",
	}

	// store appointment information into ledger
	appointmentJSON, err := json.Marshal(appointment)
	if err != nil {
		return fmt.Errorf("failed to marshal appointment: %v", err)
	}
	err = ctx.GetStub().PutState(appointmentID, appointmentJSON)
	if err != nil {
		return fmt.Errorf("failed to put appointment to ledger: %v", err)
	}

	return nil
}

func (ac *AppointmentContract) generateAppointmentID() string {
	// Increment appointment count
	ac.AppointmentCount++

	// Generate AppointmentID
	return fmt.Sprintf("apt_%d", ac.AppointmentCount)

}

// CancelAppointment cancels an existing appointment
func (ac *AppointmentContract) CancelAppointment(ctx contractapi.TransactionContextInterface, appointmentID string) error {

	// Retrieve appointment information from the ledger
	appointmentJSON, err := ctx.GetStub().GetState(appointmentID)
	if err != nil {
		return fmt.Errorf("failed to read appointment information from ledger: %v", err)
	}
	if appointmentJSON == nil {
		return fmt.Errorf("appointment not found for appointment ID %s", appointmentID)
	}

	// Unmarshal appointment information
	var appointment Appointment
	err = json.Unmarshal(appointmentJSON, &appointment)
	if err != nil {
		return fmt.Errorf("failed to unmarshal appointment information: %v", err)
	}

	// Update appointment status to "Cancelled"
	err = ac.UpdateAppointmentStatus(ctx, appointmentID, "Cancelled")
	if err != nil {
		return fmt.Errorf("failed to update appointment status: %v", err)
	}

	// Retrieve doctor availability from the ledger
	availability, err := ac.GetDoctorAvailability(ctx, appointment.DoctorID)
	if err != nil {
		return fmt.Errorf("failed to get doctor availability: %v", err)
	}

	// Add the appointment time slot back to the doctor's availability
	timeSlots, ok := availability[appointment.AppointmentDate]
	if !ok {
		return fmt.Errorf("doctor availability not found for appointment date %s", appointment.AppointmentDate)
	}
	availability[appointment.AppointmentDate] = append(timeSlots, appointment.AppointmentTime)

	// Update doctor availability on the ledger
	err = ac.UpdateDoctorAvailability(ctx, appointment.DoctorID, availability)
	if err != nil {
		return fmt.Errorf("failed to update doctor availability: %v", err)
	}

	return nil
}

// UpdateAppointmentStatus updates the status of an appointment
func (ac *AppointmentContract) UpdateAppointmentStatus(ctx contractapi.TransactionContextInterface,
	appointmentID string, status string) error {
	// Retrieve appointment information from the ledger
	appointmentJSON, err := ctx.GetStub().GetState(appointmentID)
	if err != nil {
		return fmt.Errorf("failed to read appointment information from ledger: %v", err)
	}
	if appointmentJSON == nil {
		return fmt.Errorf("appointment not found for appointment ID %s", appointmentID)
	}

	// Unmarshal appointment information
	var appointment Appointment
	err = json.Unmarshal(appointmentJSON, &appointment)
	if err != nil {
		return fmt.Errorf("failed to unmarshal appointment information: %v", err)
	}

	// Update the appointment status
	appointment.Status = status

	// Marshal the updated appointment information
	updatedAppointmentJSON, err := json.Marshal(appointment)
	if err != nil {
		return fmt.Errorf("failed to marshal updated appointment information: %v", err)
	}

	// Write the updated appointment information to the ledger
	err = ctx.GetStub().PutState(appointmentID, updatedAppointmentJSON)
	if err != nil {
		return fmt.Errorf("failed to update appointment information on ledger: %v", err)
	}

	return nil
}

// ChangeAppointment modifies the appointment time
func (ac *AppointmentContract) ChangeAppointment(ctx contractapi.TransactionContextInterface, appointmentID string,
	newAppointmentDate string, newAppointmentTime string) error {

	// Get the existing appointment information
	appointmentJSON, err := ctx.GetStub().GetState(appointmentID)
	if err != nil {
		return fmt.Errorf("failed to read appointment information from ledger: %v", err)
	}
	if appointmentJSON == nil {
		return fmt.Errorf("appointment not found for appointment ID %s", appointmentID)
	}

	// Unmarshal the existing appointment information
	var oldAppointment Appointment
	err = json.Unmarshal(appointmentJSON, &oldAppointment)
	if err != nil {
		return fmt.Errorf("failed to unmarshal appointment information: %v", err)
	}

	// Cancel the old appointment
	err = ac.CancelAppointment(ctx, appointmentID)
	if err != nil {
		return fmt.Errorf("failed to cancel old appointment: %v", err)
	}

	// Book the new appointment using the old doctor ID
	err = ac.BookAppointment(ctx, oldAppointment.DoctorID, oldAppointment.PatientID, newAppointmentDate, newAppointmentTime)
	if err != nil {
		return fmt.Errorf("failed to book new appointment: %v", err)
	}

	return nil
}

// GetAllAppointment retrieves all appointments with patient information
func (ac *AppointmentContract) GetAllAppointment(ctx contractapi.TransactionContextInterface) ([]map[string]interface{}, error) {
	// Get all keys by range to get all appointments
	appointmentIterator, err := ctx.GetStub().GetStateByRange("apt_0", "apt_999999")
	if err != nil {
		return nil, fmt.Errorf("failed to read appointments from ledger: %v", err)
	}
	defer appointmentIterator.Close()

	appointments := make([]map[string]interface{}, 0)

	for appointmentIterator.HasNext() {
		// Get the next appointment key and value
		appointmentKeyValue, err := appointmentIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate through appointments: %v", err)
		}

		// Unmarshal appointment information
		var appointment Appointment
		err = json.Unmarshal(appointmentKeyValue.Value, &appointment)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal appointment information: %v", err)
		}

		// Retrieve patient information from ledger
		patientJSON, err := ctx.GetStub().GetState(appointment.PatientID)
		if err != nil {
			return nil, fmt.Errorf("failed to read patient information from ledger: %v", err)
		}
		if patientJSON == nil {
			return nil, fmt.Errorf("patient information not found for patient ID %s", appointment.PatientID)
		}

		// Unmarshal patient information
		var patient Patient
		err = json.Unmarshal(patientJSON, &patient)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal patient information: %v", err)
		}

		// Construct appointment details
		appointmentDetails := map[string]interface{}{
			"appointmentID":   appointmentKeyValue.Key,
			"patientID":       appointment.PatientID,
			"appointmentDate": appointment.AppointmentDate,
			"appointmentTime": appointment.AppointmentTime,
			"patientName":     patient.Name,
			"patientAge":      patient.Age,
			"patientGender":   patient.Gender,
		}

		appointments = append(appointments, appointmentDetails)
	}

	return appointments, nil
}

// UpdateDoctorAvailability updates the availability of a doctor in the ledger
func (ac *AppointmentContract) UpdateDoctorAvailability(ctx contractapi.TransactionContextInterface, doctorID string, availability map[string][]string) error {
	// Marshal the updated availability to JSON
	availabilityJSON, err := json.Marshal(availability)
	if err != nil {
		return fmt.Errorf("failed to marshal updated availability: %v", err)
	}

	// Write the updated availability to the ledger
	err = ctx.GetStub().PutState(doctorID, availabilityJSON)
	if err != nil {
		return fmt.Errorf("failed to update doctor availability on ledger: %v", err)
	}

	return nil
}

// GetPatientNameByID retrieves patient's name by patient ID
func (ac *AppointmentContract) GetPatientNameByID(ctx contractapi.TransactionContextInterface, patientID string) (string, error) {
	patientJSON, err := ctx.GetStub().GetState(patientID)
	if err != nil {
		return "", fmt.Errorf("failed to read patient information from ledger: %v", err)
	}
	if patientJSON == nil {
		return "", fmt.Errorf("patient information not found for patient ID %s", patientID)
	}

	var patient Patient
	err = json.Unmarshal(patientJSON, &patient)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal patient information: %v", err)
	}

	return patient.Name, nil
}

/*
func (ac *AppointmentContract) CountAppointments(ctx contractapi.TransactionContextInterface) (int, error) {
	AppointmentIterator, err := ctx.GetStub().GetStateByRange("apt_0", "apt_999999")
	if err != nil {
		return 0, fmt.Errorf("failed to read from world state: %v", err)
	}
	defer AppointmentIterator.Close()

	totalAppointments := 0

	for AppointmentIterator.HasNext() {
		_, err := AppointmentIterator.Next()
		if err != nil {
			return 0, fmt.Errorf("failed to iterate appointment: %v", err)
		}
		totalAppointments++
	}

	return totalAppointments, nil
}
*/

func main() {
	appointmentContract := new(AppointmentContract)
	Chaincode, err := contractapi.NewChaincode(appointmentContract)
	if err != nil {
		log.Panicf("Error creating appointment chaincode %v", err)
	}

	if err := Chaincode.Start(); err != nil {
		log.Panicf("Error starting chaincode: %v", err)
	}
}
