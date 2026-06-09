
const nouns = ["quest", "hunger", "hungry", "belly", "food"];

/**
 * Called when a user issues a command to or near the mob.
 * @param {string} cmd - The command issued.
 * @param {string} rest - The arguments following the command.
 * @param {ActorObject} mob - The mob.
 * @param {RoomObject} room - The room the mob is in.
 * @param {CommandEventDetails} eventDetails - Additional event context.
 * @returns {boolean} Return true if the event was handled.
 */
function onCommand(cmd, rest, mob, room, eventDetails) {
    if (cmd == "wave") {
        mob.Command("wave");
    }
    return false;
}

/**
 * Called when a user asks the mob a question.
 * @param {ActorObject} mob - The mob.
 * @param {RoomObject} room - The room the mob is in.
 * @param {AskEventDetails} eventDetails - Details about the ask event.
 * @returns {boolean} Return true if the event was handled.
 */
function onAsk(mob, room, eventDetails) {

    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }

    match = UtilFindMatchIn(eventDetails.askText, nouns);
    if ( match.found ) {

        if ( !user.HasQuest("4-start") ) {

            mob.Command("emote rubs his belly.");
            mob.Command("say I forgot my lunch today, and I'm so hungry.");
            mob.Command("say Do you think you could find a cheese sandwich for me?");

            user.GetParty().GiveQuest("4-start");

        } else if ( user.HasQuest("4-end") ) {
            mob.Command("sayto @" + String(user.UserId()) + " Thanks, but you've done enough. Too much, really.");
        } else {
            mob.Command("emote rubs his belly.");
        }

        return true;
    }

    return false;
}

/**
 * Called when a user gives the mob an item or gold.
 * @param {ActorObject} mob - The mob.
 * @param {RoomObject} room - The room the mob is in.
 * @param {GiveEventDetails} eventDetails - Details about the give event.
 * @returns {boolean} Return true if the event was handled.
 */
function onGive(mob, room, eventDetails) {

    if (eventDetails.sourceType == "mob") {
        return false;
    }

    if ( eventDetails.gold > 0 ) {
        mob.Command("say I don't need your money... but I'll take it!");
        
        // Check a random number
        if ( Math.random() > 0.5 ) {
            mob.Command("emote flips a coin into the air and catches it!");
        } else {
            mob.Command("emote flips a coin into the air and misses the catch!");
            mob.Command("drop 1 gold");
        }
        return true;
    }

    if (eventDetails.item) {
        if (eventDetails.item.ItemId != 30004) {
            mob.Command("look !"+String(eventDetails.item.ItemId));
            mob.Command("drop !"+String(eventDetails.item.ItemId), UtilGetSecondsToTurns(5));
            return true;
        }
    }

    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }

    if ( user.HasQuest("4-start") ) {

        user.GetParty().GiveQuest("4-end");
        mob.Command("say Thanks! I can get on with my day now.");
        mob.Command("eat !"+String(eventDetails.item.ItemId) );

        return true;
    }

}


/**
 * Called each round when the mob is idle.
 * @param {ActorObject} mob - The mob.
 * @param {RoomObject} room - The room the mob is in.
 * @returns {boolean} Return true if the event was handled.
 */
// Invoked once every round if mob is idle
function onIdle(mob, room) {

    round = UtilGetRoundNumber();

    grumbled = false;
    allPlayers = room.GetPlayers();

    playersTold = mob.GetTempData('playersTold');
    if ( playersTold === null ) {
        playersTold = {};
    }

    if ( allPlayers.length > 0 ) {
        
        for( var i in allPlayers ) {

            if ( allPlayers[i].UserId() in playersTold ) {
                if ( round < playersTold[allPlayers[i].UserId()] ) {
                    continue;
                }
            }
            
            if ( !allPlayers[i].HasQuest("4-start") ) {
                if ( !grumbled ) {
                    mob.Command("emote pats his belly as it grumbles.");
                    grumbled = true;
                }
                mob.Command("sayto @" + String(allPlayers[i].UserId()) + " I'm so hungry.");
            } else {
                playersTold[allPlayers[i].UserId()] = round + 500;
            }

            playersTold[allPlayers[i].UserId()] = round + 5;
            // Don't need to repeat to every player.
            break;
        }

        if ( Object.keys(playersTold).length > 0 ) {
            mob.SetTempData('playersTold', playersTold);
        } else {
            mob.SetTempData('playersTold', null);
        }
        
        return true;
    }

    sizeBefore = Object.keys(playersTold).length;
    for (var key in playersTold) {
        if ( playersTold[key] < round-100 ) {
            delete playersTold[key];
        }
    }
    sizeAfter = Object.keys(playersTold).length;

    if ( sizeAfter != sizeBefore ) {
        if ( sizeAfter == 0 ) {
            mob.SetTempData('playersTold', null);
        } else {
            mob.SetTempData('playersTold', playersTold);
        }
    }

    action = round % 3;

    if ( action == 0 ) {
        mob.Command("wander");
        return true;
    }

    return false;
}
