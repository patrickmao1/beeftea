//define node struct, define peers list, define state, define logs, define grpc server handler

type Node struct{
	ID       int  //node id
	Peers    map[int]*Node //list of nodes
	F        int //number of faulty nodes
	LastCommit int //last commited round so we dont redo this
	Round int //current round: gets changed automatically every 6 secs

	//Inbox chan Message -channel for messages but were using gRPC
	//Store map[string]string - not sure
	//StoreMutex sync.RWMutex - i think this is a mutex lock for channels but were doing gRPC 
	
	Proposals []Proposal //all the proposals so far
	PrepareLog map[int][]Message //all the prepare messages
	CommitLog  map[int][]Message //all the commit messages

}

//message recieving starts here
func (n *Node) Run() {
	//loop infinitely and receive messages
	//only handle messages in the current round
	//call HandleMessaging method
}

func (n*Node) HandleMessaging(msg Message){
	switch msg.Type:
		case "PROPOSAL": HandleProposal(msg)
		case "PREPARE": HandlePrepare(msg)
		case "COMMIT": HandleCommit(msg)
}

//broadcast messages
func (n *Node) Broadcast(msg Message){
	for peer in Peers:
		if peer.ID != n.ID:
			//send message
}
