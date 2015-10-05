package util

import
(
	"os"
	"strings"
	"io/ioutil"
	"net"
	"time"
	"fmt"

)

// For log files, so we can restore user chat rooms and show what has been said.
const TIME_LAYOUT = "Jan 1 2015 13.12.05 -0600 GMT"

// Things we encode to send to clients so we don't break chat!
var ENCODE_UNENCODED_TOKEN = []string{"%", ":", "[", "]", ",", "\""}
var ENCODE_ENCODED_TOKEN = []string{"%25", "%3A", "%5B", "%5D", "%2C", "%22"}
var DECODE_UNENCODED_TOKEN  = []string{":", "[", "]", ",", "\"", "%"}
var DECODE_ENCODED_TOKEN = []string {"%3A", "%5B", "%5D", "%2C", "%22", "%25"}


// BEGIN STRUCTURES
// Structure for the client username and connection details, like the conneciton and room.
type Client struct
{
	// Client connection.
	userConnection net.Conn
	// Client's room, or a global room.
	Room string
	// Config file of properties.
	Prop Properties
	// Clients username
	User string
}

// Structure for logging files.
type Action struct
{
	// The commands that were previously stated.
	Comm string		`json:"comm"`
	// Action specific, such as leaving and entering a room
	Content string	`json:"content"`
	// user that performed said action
	Username string	`json:"username"`
	// ip of user
	IPAddy string	`json:"IPAddy"`
	// Timestamp of activity
	Stamp string	`json:"stamp"`

}

// This is for the config file that we can load in.
type Properties struct
{
	// Server hostname, for clients to connect to.
	Host string
	// Chat server port.
	Port string
	// Message format for when people enter and leave private rooms or connects and disconnects.
	HasEnteredRoomMsg string
	HasLeftRoomMsg string
	HasEnteredLobbyMsg string
	HasLeftLobbyMsg string

	// Format for when a person sends a message.
	ReceivedMsg string

	// Location for the JSON log file.
	LogFile string
}

// an array of actions for storage purposes to read back to user or store to log.
var actions = []Action{}

// Static client list
var curClients []*Client

// Cached version of the config file for the server.
var config = Properties{}


// END STRUCTURES
// BEGIN USEREND STUFF

// Register the connection upon connecting and cache them.
func (client *Client) Register() {
	curClients = append(curClients, client);
}

// Client closing a connection, by either exiting or just closing the window.
func (client *Client) Close(sendMessage bool) {
	if (sendMessage) {
		SendClientMessage("disconnected", "", client, false, client.Prop)
	}
	// Close the connection to client and remove from user list.
	client.userConnection.Close();
	curClients = removeEntry(client, curClients);
}

// Remove a client from the registry of users.
func removeEntry(client *Client, clientArr []*Client) []*Client {
	// Our return value after removing the client.
	retVal := clientArr
	// Need to declare index, if not found then we return the same array.
	// This is why it's negative one.
	index := -1
	for i, curVal := range clientArr {
		// If the current value is the client, set index to i, then we continue with removing.
		if(curVal == client) {
			index = i;
			break;
		}
		// If nothing is found we have to just iterate through the entire client
		// array, then simply return the same array.
	}

	// So that being said, if we have the user, we must remove.
	if(index >= 0) {
		// Initialize the new array, make it one less the current length.
		retVal = make([]*Client, len(clientArr)-1)
		// Copy over the all the clients, up to the client that needs to be removed.
		copy(retVal, clientArr[:index])
		// Now, we copy over the rest of the array, because it's removing that user.
		copy(retVal[index:], clientArr[index+1:])
		// This way, we completely ignore the user that's in the old list.
		// This may cause some issues depending on the amount of users in the list.
		// Could this lead to memory errors possibly?	
	}
	// Return the new user list.
	return retVal;
}

// Function to send client messages to either users, or channels, etc.
func SendClientMessage(msgType string, msg string, client *Client, thisOneOnly bool, property Properties) {
	// Check if the message if only for the provided client, for server commands.
	if (thisOneOnly) {
		msg = fmt.Sprintf("/%v", msgType);
		fmt.Fprintln(client.userConnection, msg)
	} else if (client.User != "") {
		// Now if it isn't send message to everyone.
		// Log the action.
		LogAction(msgType, msg, client, property);

		// Construct payload to be sent to clients.
		pLoad := fmt.Sprintf("/%v [%v] %v", msgType, client.User, msg);

		for _, _client := range curClients {
			// Write the emssage to the client.
			if ((thisOneOnly && _client.User == client.User) || 
			(!thisOneOnly && _client.User != "")) {
				// Only see the message if in the same room as each other.
				if (msgType == "message" && client.Room != _client.Room) {
					continue;
				}
			}
			fmt.Fprintln(_client.userConnection, pLoad)
		}
	}
}
// END USEREND STUFF

//BEGIN MISC

// Simple error checker to see if messages are empty.
func CheckForError(errNo error, msg string) {
	if errNo != nil {
		println(msg + ": ", errNo.Error())
		os.Exit(1)
	}
}

// Use a function to double quote single quotes to look nice.
func EncodeCSV(val string) (string) {
	return strings.Replace(val, "\"", "\"\"", -1)
}

// End point encoding to help with special characters.
func Encode(val string) (string) {
	return replace(ENCODE_UNENCODED_TOKEN, DECODE_UNENCODED_TOKEN, val)
}

// End point for replacing special characters.
func Decode (val string) (string) {
	return replace(DECODE_ENCODED_TOKEN, DECODE_UNENCODED_TOKEN, val)
}

// Replace function for encoding.
func replace(tokenReplacers []string, toTokens []string, val string) (string) {
	for i := 0; i < len(tokenReplacers); i++ {
		val = strings.Replace(val, tokenReplacers[i], toTokens[i], -1)
	}
	return val;
}

func LogAction(act string, msg string, client *Client, property Properties) {
	// Get the IP and timestamp for it to be logged.
	ipAddy := client.userConnection.RemoteAddr().String()
	stampOfTime := time.Now().Format(TIME_LAYOUT)

	// Keep track of all actions in the action array.
	actions = append(actions, Action {
		Comm: act,
		Content: msg,
		Username: client.User,
		IPAddy: ipAddy,
		Stamp: stampOfTime,
	})
	// If the logfile actually exists.
	if (property.LogFile != "") {
		// If the message is nothing.
		if(msg == "") {
			msg = "N/A"
		}
		fmt.Printf("Logging values %s, %s, %s\n", act, msg, client.User);
		
		messageLog := fmt.Sprintf("\"%s\", \"%s\", \"%s\", \"%s\", \"%s\"\n",
			EncodeCSV(client.User), EncodeCSV(act), EncodeCSV(msg), EncodeCSV(stampOfTime),
			EncodeCSV(ipAddy))

		// Open the log file, check for errors.
		file, errNo := os.OpenFile(property.LogFile, os.O_APPEND|os.O_WRONLY, 0600)
		if (errNo != nil) {
			// Try and create the file first. If not, we open.
			errNo = ioutil.WriteFile(property.LogFile, []byte{}, 0600)
			file, errNo = os.OpenFile(property.LogFile, os.O_APPEND|os.O_WRONLY, 0600)
			CheckForError(errNo, "Cannot create log file!")
		}
		// Close the file.
		defer file.Close()
		_, errNo = file.WriteString(messageLog)
		CheckForError(errNo, "Cannot write to log file.")
	}
}

func LoadConfig() Properties {
	// Check if port is not empty, simple error check.
	if(config.Port != "") {
		return config;
	}
	// Get our current directory.
	// curWD, _ := os.Getwd()

	// Load the config file here, so we can access and setup the server.
	// confFile, errNo := ioutil.ReadFile(curWD + "/config.json")
	// CheckForError(errNo, "JSON File cannot be read.")
	// JSON library is required here to read from the file. Debating whether we should do
	// or not, or just hard encode the port and messages. Maybe do that.
	// For now, will put implementation but will comment out until for sure we are allowed
	// to use.
	// var confData map[string]interface{}
	// errNo = json.Unmarshal(confFile, &confData)
	// CheckForError(errNo, "Invliad JSON File.")
	// var rturnVals = Properties
	// {
	//	Host: confData["Host"].(string),
	//	Port: confData["Port"].(string),
	//	HasEnteredRoomMsg: confData["HasEnteredRoomMsg"].(string),
	//	HasLeftRoomMsg: confData["HasLeftRoomMsg"].(string),
	//	HasEnteredLobbyMsg: confData["HasEnteredLobbyMsg"].(string),
	//	HasLeftLobbyMsg: confData["HasLeftLobbyMsg"].(string),
	//  ReceivedMsg: confData["ReReceivedMsg"].(string),
	//	LogFile: confData["LogFile"].(string),
	// }

	var rturnVals = Properties {
		Host: "localhost",
		Port: "25565",
		HasEnteredRoomMsg: "[%s] has entered the room \"%s\"",
		HasLeftRoomMsg: "[%s] has left the room \"%s\"",
		HasEnteredLobbyMsg: "[%s] has entered the lobby",
		HasLeftLobbyMsg: "[%s] has left the lobby",
		ReceivedMsg: "[%s] says: %s",
		LogFile: "./log.txt",
	}
	config = rturnVals;
	return rturnVals;
}

