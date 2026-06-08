
mapSignData = "";

/**
 * Called when a user enters the room.
 * @param {ActorObject} user - The user entering the room.
 * @param {RoomObject} room - The room being entered.
 * @returns {boolean} Return false to suppress the automatic look.
 */
function onEnter(user, room) {
    // Special case for if the player left the game while in jail.
    // The ephemeral room gets destroyed and the player gets sent back to TS
    // From here we can put them back in jail.
    if ( user.TimerExists("jail") ) {
        user.MoveRoom(1003);
        return false;
    }
    return true;
}

/**
 * Called when a user issues a command in the room.
 * @param {string} cmd - The command issued.
 * @param {string} rest - The arguments following the command.
 * @param {ActorObject} user - The user issuing the command.
 * @param {RoomObject} room - The room where the command was issued.
 * @returns {boolean} Return true if the command was handled.
 */
function onCommand(cmd, rest, user, room) {
    if (cmd != "look" && cmd != "read" ) {
        return false;
    }
    
    if ( rest.substr(rest.length - 3) == "map" || rest.substr(rest.length - 4) == "sign" ) {
      
        SendUserMessage(user.UserId(), "You look at the map nailed to the sign.");
        SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" looks at the map nailed to the sign.", user.UserId());

        // Load the cached map, or re-generate and cache it if it's not there
        if ( mapSignData == "" ) {
            mapSignData = GetMap(room.RoomId(), 1, 22, 38, "Map of Frostfang", false, String(room.RoomId())+",×,Here");
        }

        // Send the map to the user.
        SendUserMessage(user.UserId(), mapSignData);

        return true;
    }
    
    return false;
}

/**
 * Called when the room first loads.
 * @param {RoomObject} room - The room that loaded.
 * @returns {void}
 */
function onLoad(room) {
    // Just running this to pre-cache the map so that if someone looks at the map it won't time out
    mapSignData = GetMap(room.RoomId(), 1, 22, 38, "Map of Frostfang", false, String(room.RoomId())+",×,Here");
}
