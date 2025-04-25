//check vrf eligible, check vrf score

//returns if the vrf is eligible to propose in this round
func Eligible(id int, round int) bool {
	// need to pick a threshold number, hash our id and check if its less than threshold
}


//returns a score for a given node/round combo
func Score(id int, round int) float64 {
	//determinisitc hashed score from VRF output
	return HMACtoFloat(id, round)
}

//need to change this up this is just filler code from CHAT 

//need the algorand psuedo code
