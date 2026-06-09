
/**
 * Called when the buff is first applied to the actor.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onStart(actor, triggersLeft) {

    if ( actor.HasBuffFlag("warmed")  ) {
        actor.RemoveBuff(31);
        return;
    }
    harmAmt = actor.AddHealth(-1 * UtilDiceRoll(1, 2));
    if (harmAmt < 1 ) {
        harmAmt *= -1;
    }
    SendUserMessage(actor.UserId(),     '<ansi fg="51">The cold bites for <ansi fg="damage">'+String(harmAmt)+' damage</ansi>!</ansi>\n');
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' is freezing.', actor.UserId());


}