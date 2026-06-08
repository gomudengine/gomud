
/**
 * Called when the buff is first applied to the actor.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onStart(actor, triggersLeft) {
    
    quarryUserName = actor.GetMiscCharacterData("tracking-user");
    quarryMobName = actor.GetMiscCharacterData("tracking-mob");

    if ( quarryUserName != null ) {
        SendUserMessage(actor.UserId(), 'Your senses are heightened as you focus your tracking skills on <ansi fg="username">'+quarryUserName+'</ansi>.');
    } else {
        SendUserMessage(actor.UserId(), 'Your senses are heightened as you focus your tracking skills on <ansi fg="mobname">'+quarryMobName+'</ansi>.');
    }

}

/**
 * Called when the buff expires or is removed.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onEnd(actor, triggersLeft) {

    quarryUserName = actor.GetMiscCharacterData("tracking-user");
    quarryMobName = actor.GetMiscCharacterData("tracking-mob");

    if ( quarryUserName != null ) {
        SendUserMessage(actor.UserId(), 'You are no longer actively tracking <ansi fg="username">'+quarryUserName+'</ansi>.');
    } else {
        SendUserMessage(actor.UserId(), 'You are no longer actively tracking <ansi fg="mobname">'+quarryMobName+'</ansi>.');
    }

    actor.SetMiscCharacterData("tracking-mob", null);
    actor.SetMiscCharacterData("tracking-user", null);

    

}
