

const trapNouns = ["trap", "rat trap", "rat", "rodric"];

/**
 * Called when a user asks the mob a question.
 * @param {ActorObject} mob - The mob.
 * @param {RoomObject} room - The room the mob is in.
 * @param {object} eventDetails - Details about the ask event (sourceId, askText).
 * @returns {boolean} Return true if the event was handled.
 */
function onAsk(mob, room, eventDetails) {

    user = GetUser(eventDetails.sourceId);

    if ( user.HasQuest("7-gettrap") && !user.HasQuest("7-tradetrap") ) {

        match = UtilFindMatchIn(eventDetails.askText, trapNouns);
        if ( match.found ) {

            mob.Command("say Rodric asked for his rat trap back? I've been waiting for him to pick this up for weeks.");
            item = CreateItem(11);
            
            mob.GiveItem(item);
            mob.Command("give !11 @" + String(eventDetails.sourceId));

            mob.Command("say Thanks for picking it up!");

            user.GetParty().GiveQuest("7-tradetrap");

            return true;
        }

        return false;
    }

    return false;
}
