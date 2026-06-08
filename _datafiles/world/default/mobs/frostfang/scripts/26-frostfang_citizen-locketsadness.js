
const sadnessSubjects = ["quest", "locket", "sad", "sadness", "crying", "sniffles", "necklace"];
const gardenSubjects = ["garden", "where", "gardening", "quest", "locket", "sad", "sadness", "necklace"];

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

    if ( user.HasQuest("1-end") ) {

        match = UtilFindMatchIn(eventDetails.askText, sadnessSubjects);
        if ( match.found ) {

            mob.Command("say I'm so absent minded sometimes. I'm glad you found it.");

            return true;
        }

        return false;
    }

    if ( !user.HasQuest("1-start") ) {

        match = UtilFindMatchIn(eventDetails.askText, sadnessSubjects);

        if ( match.found ) {

            mob.Command("emote sighs deeply.");
            mob.Command("say I lost my locket. I think it was when I was gardening.");

            user.GetParty().GiveQuest("1-start");

            return true;
        }

    } else {

        match = UtilFindMatchIn(eventDetails.askText, gardenSubjects);
        if ( match.found ) {
            mob.Command("emote thinks hard for a moment.");
            mob.Command("say I was trimming the hedges out back the last time I remember wearing it.");

            return true;
        }

    }

    return true;
}

/**
 * Called when a user gives the mob an item or gold.
 * @param {ActorObject} mob - The mob.
 * @param {RoomObject} room - The room the mob is in.
 * @param {GiveEventDetails} eventDetails - Details about the give event.
 * @returns {boolean} Return true if the event was handled.
 */
function onGive(mob, room, eventDetails) {
    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }
    
    
    showLocketCounter = mob.GetTempData('showLocketCounter');
    if ( showLocketCounter === null ) {
        showLocketCounter = {};
    }

    if (eventDetails.item) {

        if (eventDetails.item.ItemId == 20025) {
            
            mob.Command("say Thank you so much! I thought I'd never see this again!");

            if ( !showLocketCounter[eventDetails.sourceId] ) {
                showLocketCounter[eventDetails.sourceId] = 0;
            }

            // Give it back to them
            if ( showLocketCounter[eventDetails.sourceId] > 2 ) {
                mob.GiveItem(20033); // Spawn the item in the mobs posession
                mob.Command("give !20033 @" + String(eventDetails.sourceId)); // Give it to the player using shorthand
            }
            
            user.GetParty().GiveQuest("1-end");

            return true;

        }

        if ( !user.HasQuest("1-end") ) {
            mob.Command("say Thank you, but nothing could ever replace my locket.");
            // Give it back to them
            mob.Command("give !"+eventDetails.item.ItemId+" @" + String(eventDetails.sourceId));
        }

        return true;
    }

    if ( eventDetails.gold > 0 ) {
        mob.Command("say Just what kind of girl do you think I am???");
        return true;
    }

    return false;
}



/**
 * Called when a user shows the mob an item.
 * @param {ActorObject} mob - The mob.
 * @param {RoomObject} room - The room the mob is in.
 * @param {ShowEventDetails} eventDetails - Details about the show event.
 * @returns {boolean} Return true if the event was handled.
 */
function onShow(mob, room, eventDetails) {
    showLocketCounter = mob.GetTempData('showLocketCounter');
    if ( showLocketCounter === null ) {
        showLocketCounter = {};
    }

    if (eventDetails.item.ItemId == 20025) {
        
        if ( !showLocketCounter[eventDetails.sourceId] ) {
            showLocketCounter[eventDetails.sourceId] = 0;
        }
        showLocketCounter[eventDetails.sourceId]++;

        mob.SetTempData('showLocketCounter', showLocketCounter);

        if ( showLocketCounter[eventDetails.sourceId] == 1 ) {
            mob.Command("say Wow, that's it! Can I have it back?");
            return true;
        }

        if ( showLocketCounter[eventDetails.sourceId] == 2 ) {
            mob.Command("say Please, it's only worth is sentimental value. Can I have it back?");
            return true;
        }

        if ( showLocketCounter[eventDetails.sourceId] > 2 ) {
            mob.Command("say I can trade you for this other locket I have of equal value.");
            return true;
        }

        return true;
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

    switch (UtilGetRoundNumber() % 4) {
    case 0:
        mob.Command("emote looks sad.");
        return true;
    case 2:
        mob.Command("emote sniffles a bit, holding back tears.");
        return true;
    default: // 1, 3
        if ( UtilDiceRoll(1, 10) == 1 ) {
            mob.Command("pathto 274"); // look around the bushes area
            return true;
        }
        return false;
    }

}