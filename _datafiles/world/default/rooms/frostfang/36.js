

/**
 * Called when a user issues a north command in the room.
 * @param {string} rest - The arguments following the command.
 * @param {ActorObject} user - The user issuing the command.
 * @param {RoomObject} room - The room where the command was issued.
 * @returns {boolean} Return true if the command was handled.
 */
function onCommand_north(rest, user, room) {


    hasQuestUserIds = room.HasQuest("2-start", user.UserId());
    if ( hasQuestUserIds.length < 1 ) {

        SendUserMessage(user.UserId(), 'The guards block your path. "<ansi fg="yellow">You must be invited to enter the throne room</ansi>," they say. "<ansi fg="yellow">We cannot let you pass.</ansi>"');
        SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+' tries to enter the throne room, but the guards block the way.', user.UserId());
        
        return true;
    }

    return false;

}