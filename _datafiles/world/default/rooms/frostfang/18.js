
const altar = ["altar"];

/**
 * Called when a user issues a look command in the room.
 * @param {string} rest - The arguments following the command.
 * @param {ActorObject} user - The user issuing the command.
 * @param {RoomObject} room - The room where the command was issued.
 * @returns {boolean} Return true if the command was handled.
 */
function onCommand_look(rest, user, room) {

    matches = UtilFindMatchIn(rest, altar);
    if ( matches.found ) {
        SendUserMessage(user.UserId(), "<ansi fg=\"240\">The smell of the insense fills your nostrels, numbing your senses.</ansi>");       
        user.GiveBuff(2, "drugs");
        return true;
    }

    return false;
}
