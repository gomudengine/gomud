
const JAIL_TIME = "1 hour";

/**
 * Called when a user enters the room.
 * @param {ActorObject} user - The user entering the room.
 * @param {RoomObject} room - The room being entered.
 * @returns {boolean} Return false to suppress the automatic look.
 */
function onEnter(user, room) {

    user.SetHealth(1);
    
    if ( !room.IsEphemeral() ) {

        var newRoomIds = CreateInstancesFromRoomIds( [room.RoomId()] );

        if ( newRoomIds[room.RoomId()] ) {
            user.MoveRoom(newRoomIds[room.RoomId()]);
            return false;
        } 

    }
    
    if ( !user.TimerExists("jail") ) {

        user.AddEventLog(`jail`, `Thrown in jail`);

        user.TimerSet("jail", JAIL_TIME);

        room.SendText("");
        room.SendText("<ansi fg=\"red-bold\">********************************************************************************</ansi>");
        room.SendText("You hear a loud <ansi fg=\"red-bold\">!!!CLANK!!!</ansi>, and can immediately tell...");
        room.SendText("The cell door is LOCKED from the other side!");
        room.SendText('You hear someone shout, <ansi fg="saytext-mob">"Maybe an hour in a cell will cool you off!"</ansi>');
        room.SendText("<ansi fg=\"red-bold\">********************************************************************************</ansi>");
        room.SendText("");
        
        user.Command("look", 1);
        
    }

    return false;
}

/**
 * Called each round when the room is idle (no players or on a timer).
 * @param {RoomObject} room - The room.
 * @returns {boolean} Return true if the event was handled.
 */
function onIdle(room) {

    if ( room.IsLocked("cell door") ) {
        var playersInRoom = room.GetPlayers();
        for( var i in playersInRoom ) {
            if ( playersInRoom[i].TimerExpired("jail") ) {
                room.SendText("You hear a loud <ansi fg=\"red-bold\">!!!KA-LUNK!!!</ansi>, and the cell door is UNLOCKED from the other side.");
                room.SetLocked("cell door", false);
                playersInRoom[i].AddEventLog(`jail`, `Released from jail`);
            }
        }
    }

    return true;
}