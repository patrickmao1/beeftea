//proposal phase, pBFT consensus, round timer, logging, transitions
//func startRound, func HandlePrepare, func HandleCommit

//automatically increments to next round every 6 seconds
func (n *Node) StartRoundTimer(){
	//every 6 seconds:
		n.Round += 1
		ResetRoundState()
		StartProposalPhase()
}

//clear all the logs every round
func (n * Node) ResetRoundStste() {
	Proposals = []
	delete PrepareLog[n.round]
	delete CommitLog[n.Round]
}

//Proposal phase - this should be the first two secs of a round

func (n *Node) StartProposalPhase(){
	if Eligible(n.ID, n.Round): //check if vrf eligible
		op := GetNextOperation() //get the operation from proposer
		score := Score(n.ID, n.Round) //set the score usingt the proposers score

		// not sure what goes here proposal := Proposal{}
		// not sure here msg := Message{Type: "PROPOSAL", score, }
		Broadcast(msg) //call broadcast message

}

func (n *Node) StartPBFTPhase(){
	proposal := SelectLowestProposal()
	prepare := Message{Type:"PREPARE", Operation: proposal.Operation}
	Broadcast(prepare)
}

func (n *Node) HandleProposal(msg Message){
	//store valid proposal
	//get proposal: proposal := Proposal{}
	Proposals.append(proposal)
}

func (n *Node) HandlePrepare(msg Message) {
	Append to PrepareLog[msg.Round]
	if CountMatching(PrepareLog[msg.Round]) >= 2*F:
		Broadcast(COMMIT)

}

func(n *Node) HandleCommit(msg Message){
	Append to CommitLog[msg.Round]
	if CountMatching(CommitLog[msg.Round]) >= 2*f:
	ApplyOperation(msg.Operation)
}

