
/**
 * Called when a user issues a west command in the room.
 * @param {string} rest - The arguments following the command.
 * @param {ActorObject} user - The user issuing the command.
 * @param {RoomObject} room - The room where the command was issued.
 * @returns {boolean} Return true if the command was handled.
 */
function onCommand_west(rest, user, room) {

    if ( !UtilIsDay() ) {
        SendUserMessage(user.UserId(), "The eastern city gates close every night. You'll have to wait for day, or find another way in.");
        return true;
    }
    
    return false;
}
