
/**
 * Called when a user issues an east command in the room.
 * @param {string} rest - The arguments following the command.
 * @param {ActorObject} user - The user issuing the command.
 * @param {RoomObject} room - The room where the command was issued.
 * @returns {boolean} Return true if the command was handled.
 */
function onCommand_east(rest, user, room) {

    if ( !UtilIsDay() ) {
        SendUserMessage(user.UserId(), "The east gates are closed for the night. You'll have to wait for day, or find another way out.");
        return true;
    }
    
    return false;
}
