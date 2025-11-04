

const nouns = ["sarcophagus", "tomb"];
const verbs = ["touch", "push", "hit", "kick", "open", "pry"];


// Generic Command Handler
function onCommand(cmd, rest, user, room) {

    if ( !verbs.includes(cmd) ) {
        return false;
    }
    
    matches = UtilFindMatchIn(rest, nouns);
    if ( !matches.found ) {
        return false;
    }

    SendUserMessage(user.UserId(), "The room begins to tremble, and a trap door opens beneath your feet! You fall into the room below!");
    SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" has triggered some sort of trap! the room begins to tremble and a trap door opens beneath your feet. You fall into the darkness below.", user.UserId());


    players = room.GetPlayers();
    for( var i in players ) {
        SendRoomMessage(138, players[i].GetCharacterName(true)+" falls into the room from above.", players[i].UserId());
        players[i].MoveRoom(138);
    }
    
    return true;
}


      