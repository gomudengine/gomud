
const verbs = ["roll", "push"];
const nouns = ["boulder", "rock"];

/**
 * Called when a user issues a command in the room.
 * @param {string} cmd - The command issued.
 * @param {string} rest - The arguments following the command.
 * @param {ActorObject} user - The user issuing the command.
 * @param {RoomObject} room - The room where the command was issued.
 * @returns {boolean} Return true if the command was handled.
 */
function onCommand(cmd, rest, user, room) {

    if ( !verbs.includes(cmd) ) {
        return false;
    }
    
    matches = UtilFindMatchIn(rest, nouns);
    if ( !matches.found ) {
        return false;
    }

    if ( room.HasMutator('pushed-boulder') ) {
        SendUserMessage(user.UserId(), "The boulder is already pushed aside.");
        return true;
    }

    SendUserMessage(user.UserId(), "You roll the boulder to the side, revealing a pathway!");

    room.AddMutator('pushed-boulder');
        
    return true;
}

