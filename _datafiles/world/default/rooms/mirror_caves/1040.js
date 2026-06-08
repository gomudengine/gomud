/**
 * Called when a user issues an out command in the room.
 * @param {string} rest - The arguments following the command.
 * @param {ActorObject} user - The user issuing the command.
 * @param {RoomObject} room - The room where the command was issued.
 * @returns {boolean} Return true if the command was handled.
 */
function onCommand_out(rest, user, room) {
    
    var mobs = room.GetMobs();
    if ( mobs.length > 0 ) {
        SendUserMessage(user.UserId(), "The way out is block by denizens of the cave.");
        return true;
    }

    return false;
}
