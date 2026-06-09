
/**
 * Called when a user issues a grate command in the room.
 * @param {string} rest - The arguments following the command.
 * @param {ActorObject} user - The user issuing the command.
 * @param {RoomObject} room - The room where the command was issued.
 * @returns {boolean} Return true if the command was handled.
 */
function onCommand_grate(rest, user, room) {

    SendUserMessage(user.UserId(), "You take a deep breath and wade through the filth.");

    return false;
}
