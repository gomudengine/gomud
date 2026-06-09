
crowbarAvailableRound = 0;

const crowbar = ["crowbar", "rod", "metal", "bar"];
const verbs = ["get", "take", "grab", "steal", "snatch"];


/**
 * Called when a user enters the room.
 * @param {ActorObject} user - The user entering the room.
 * @param {RoomObject} room - The room being entered.
 * @returns {boolean} Return false to suppress the automatic look.
 */
function onEnter(user, room) {
    user.GiveBuff(15, "sleep");
    return false; // return false to prevent the "auto look"
}

/**
 * Called when a user issues a look command in the room.
 * @param {string} rest - The arguments following the command.
 * @param {ActorObject} user - The user issuing the command.
 * @param {RoomObject} room - The room where the command was issued.
 * @returns {boolean} Return true if the command was handled.
 */
function onCommand_look(rest, user, room) {

    matches = UtilFindMatchIn(rest, crowbar);
    if ( !matches.found ) {
        return false;
    }

    roundNow = UtilGetRoundNumber();

    if (roundNow < crowbarAvailableRound) {
        return false;
    }

    SendUserMessage(user.UserId(), "A <ansi fg=\"item\">crowbar</ansi> leans besides the fireplace. Probably used for poking the fire and moving the logs around.");
    SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" looks at the <ansi fg=\"item\">crowbar</ansi> by the fireplace.", user.UserId());   

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

    if ( !verbs.includes(cmd) ) {
        return false;
    }
    
    matches = UtilFindMatchIn(rest, crowbar);
    if ( !matches.found ) {
        return false;
    }

    roundNow = UtilGetRoundNumber();
    
    if (roundNow < crowbarAvailableRound) {
        return false;
    }

    crowbarAvailableRound = roundNow + UtilGetMinutesToRounds(15);

    SendUserMessage(user.UserId(), "You take the <ansi fg=\"item\">crowbar</ansi>. They probably won't miss it.");
    SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" takes the <ansi fg=\"item\">crowbar</ansi> from beside the fireplace.", user.UserId());   
    
    user.GiveItem(10012);

    return true;
}


