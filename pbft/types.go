//proposal struct(num for vrf)
//message struct(prepare, commit)

type Message struct{
	Type string // it will be either "PROPOSAL", "COMMIT", "PREPARE"
	Round int
	Operation string //we will add or delete from the key val store
	SenderID int

}

//proposal message sent during proposal phase
type Proposal struct{
	Round int
	VRFScore float64 //the lower vrf is picked
	operation string
	ProposerID int
}

//operation to apple to the key val store
type Operation struct{
	Type string 
	Key string
	Value string
}