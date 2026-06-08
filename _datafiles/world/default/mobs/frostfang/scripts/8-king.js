
const startQuestSubjects = ["quest", "bishop",  "arch", "arch-bishop", "archbishop", "trust"];
const lichSubjects = ["lich", "old king", "evil king", "tomb", "sarcophagus"];

/**
 * Called when a user asks the mob a question.
 * @param {ActorObject} mob - The mob.
 * @param {RoomObject} room - The room the mob is in.
 * @param {object} eventDetails - Details about the ask event (sourceId, askText).
 * @returns {boolean} Return true if the event was handled.
 */
function onAsk(mob, room, eventDetails) {

    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }

    match = UtilFindMatchIn(eventDetails.askText, startQuestSubjects); 
    if ( match.found ) {

        if ( user.HasQuest("2-start") ) {
            mob.Command("say Maybe you could snoop around there a bit and see if you can discover anything. They are just south of Town Square.");
            return false;
        }

        mob.Command("say Don't let the Sanctuary of the Benevolent Heart fool you... they are up to something.");
        mob.Command("say My spies haven't been able to discover anything suspicious about their behavior, which is the first clue something is up.");
        mob.Command("say Maybe you could snoop around there a bit and see if you can discover anything. They are just to the south of Town Square.");
        
        user.GetParty().GiveQuest("2-start");

        return true;
    }

    match = UtilFindMatchIn(eventDetails.askText, lichSubjects);
    if ( match.found ) {
        mob.Command("say An ancient lich king eh? Do you have any proof that what you say is true?");

        return true;
    }

    return true;
}

/**
 * Called when a user gives the mob an item or gold.
 * @param {ActorObject} mob - The mob.
 * @param {RoomObject} room - The room the mob is in.
 * @param {object} eventDetails - Details about the give event (sourceId, sourceType, item, gold).
 * @returns {boolean} Return true if the event was handled.
 */
function onGive(mob, room, eventDetails) {
    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }    

    if (eventDetails.item) {
        if (eventDetails.item.ItemId == 20018) {
            
            mob.Command("say Thank you for taking care of that problem. The kingdom is indebted to you.");
            mob.Command("say I will add this artifact to the treasury. Here is some gold to compensate you.");

            mob.AddGold(1250);
            mob.Command("give 1250 gold @" + String(eventDetails.sourceId));

            user.GetParty().GiveQuest("2-end");

            return true;
        } else {
            mob.Command("say I have no need for that.");
            // Give it back to them
            mob.Command("give !"+String(eventDetails.item.ItemId) + " @" + String(eventDetails.sourceId));
        }
        return true;
    }
    return false;
}

/**
 * Called when a user shows the mob an item.
 * @param {ActorObject} mob - The mob.
 * @param {RoomObject} room - The room the mob is in.
 * @param {object} eventDetails - Details about the show event (sourceId, item).
 * @returns {boolean} Return true if the event was handled.
 */
function onShow(mob, room, eventDetails) {

    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }
    
    if (eventDetails.item.ItemId == 20018) {
        
        mob.Command("say Thank you for taking care of that problem. The kingdom is indebted to you.");

        user.GetParty().GiveQuest("2-end");

        return true;

    } else {
        mob.Command("nods patronizingly.");
    }
    
    return false;
}


/**
 * Called each round when the mob is idle.
 * @param {ActorObject} mob - The mob.
 * @param {RoomObject} room - The room the mob is in.
 * @returns {boolean} Return true if the event was handled.
 */
function onIdle(mob, room) {

    round = UtilGetRoundNumber();
    action = round % 3;

    if (action > 0) {
        return false;
    }

    mob.Command("emote grumbles under his breath");

    missingQuestUsers = room.MissingQuest("2-start");
    if ( missingQuestUsers.length > 0 ) {
        mob.Command("say I really don't trust the arch-bishop.");
    }

    return true;
}