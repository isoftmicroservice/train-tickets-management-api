package model

type User struct {
	FirstName string
	LastName  string
	Email     string
}

type Seat struct {
	Section    string
	SeatNumber int32 
}

type Ticket struct {
	From      string 
	To        string 
	User      User
	PricePaid int32 
	Seat      Seat
}

func IsValidSection(section string) bool {
	return section == "A" || section == "B"
}

func IsValidSeatNumber(seatNumber int32) bool {
	return seatNumber >= 1 && seatNumber <= 10
}

