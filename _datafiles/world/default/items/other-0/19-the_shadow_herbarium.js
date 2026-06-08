
/**
 * Called when a user issues a read command on the item.
 * @param {ActorObject} user - The user issuing the command.
 * @param {ItemObject} item - The item.
 * @param {RoomObject} room - The room where the command was issued.
 * @returns {boolean} Return true if the command was handled.
 */
function onCommand_read(user, item, room) {

    SendUserMessage(user.UserId(), "You thumb through your <ansi fg=\"item\">"+item.Name(true)+"</ansi> book.");
    SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" thumbs through their <ansi fg=\"item\">"+item.Name(true)+"</ansi> book.", user.UserId());   

    if ( user.LearnSpell("curepoison") ) {
        SendUserMessage(user.UserId(), "You discover the the <ansi fg=\"spell-helpful\">Cure Poison</ansi> spell. It can remove a deadly ailment.");
        SendUserMessage(user.UserId(), "Check your <ansi fg=\"command\">spellbook</ansi>.");
        SendUserMessage(user.UserId(), "The book disinigrates in your hands.");
        item.SetUsesLeft(0);
    }

    return true;
}


/**
 * Called when a user issues a use command on the item.
 * @param {ActorObject} user - The user issuing the command.
 * @param {ItemObject} item - The item.
 * @param {RoomObject} room - The room where the command was issued.
 * @returns {boolean} Return true if the command was handled.
 */
function onCommand_use(user, item, room) {
    return onCommand_read(user, item, room);
}