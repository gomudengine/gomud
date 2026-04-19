
// Major Arcana: name, upright meaning, reversed meaning
var MAJOR_ARCANA = [
    { name: "The Fool",         upright: "new beginnings, innocence, and a leap into the unknown",          reversed: "recklessness, poor judgment, and a step taken without looking" },
    { name: "The Magician",     upright: "willpower, skill, and the ability to manifest your desires",       reversed: "manipulation, untapped potential, and misdirected energy" },
    { name: "The High Priestess", upright: "intuition, mystery, and hidden knowledge",                      reversed: "secrets withheld, surface-level understanding, and ignored instincts" },
    { name: "The Empress",      upright: "abundance, nurturing, and creative fertility",                     reversed: "dependence, stagnation, and creative blocks" },
    { name: "The Emperor",      upright: "authority, structure, and a steady hand",                          reversed: "rigidity, domination, and an unwillingness to bend" },
    { name: "The Hierophant",   upright: "tradition, conformity, and the wisdom of institutions",            reversed: "rebellion, unconventional paths, and questioning the rules" },
    { name: "The Lovers",       upright: "union, harmony, and choices made from the heart",                  reversed: "disharmony, imbalance, and a decision you may regret" },
    { name: "The Chariot",      upright: "determination, control, and victory through willpower",            reversed: "aggression, lack of direction, and a runaway course" },
    { name: "Strength",         upright: "courage, patience, and quiet inner power",                         reversed: "self-doubt, weakness, and energy turned against itself" },
    { name: "The Hermit",       upright: "solitude, introspection, and the light found in stillness",        reversed: "isolation, loneliness, and wisdom hoarded rather than shared" },
    { name: "Wheel of Fortune", upright: "cycles, fate, and a turn of luck in your favor",                   reversed: "bad luck, resistance to change, and a wheel spinning out of control" },
    { name: "Justice",          upright: "fairness, truth, and the weight of consequence",                   reversed: "dishonesty, injustice, and scales that have been tipped" },
    { name: "The Hanged Man",   upright: "surrender, new perspective, and wisdom found in waiting",          reversed: "stalling, resistance, and a refusal to let go" },
    { name: "Death",            upright: "endings, transformation, and the door that must be walked through", reversed: "resistance to change, decay, and clinging to what is already gone" },
    { name: "Temperance",       upright: "balance, moderation, and the patience of a long game",             reversed: "excess, imbalance, and a lack of long-term vision" },
    { name: "The Devil",        upright: "bondage, materialism, and the chains you have chosen",             reversed: "release, reclaiming power, and breaking free from what binds you" },
    { name: "The Tower",        upright: "sudden upheaval, revelation, and the collapse of false structures", reversed: "fear of change, averting disaster, and a crisis narrowly avoided" },
    { name: "The Star",         upright: "hope, renewal, and the calm after a long storm",                   reversed: "despair, lost faith, and a light that has gone dim" },
    { name: "The Moon",         upright: "illusion, fear, and the things that stir in the dark",             reversed: "confusion lifting, repressed emotion surfacing, and lies coming to light" },
    { name: "The Sun",          upright: "joy, vitality, and the warmth of a clear day",                     reversed: "temporary setbacks, overconfidence, and joy that hasn't quite arrived yet" },
    { name: "Judgement",        upright: "reflection, reckoning, and a call you cannot ignore",              reversed: "self-doubt, refusal to learn, and a reckoning postponed" },
    { name: "The World",        upright: "completion, integration, and the satisfaction of a journey's end", reversed: "incompletion, loose ends, and a destination not yet reached" },
];

var SHUFFLE_MESSAGES = [
    "You riffle the cards with practiced ease, the deck whispering softly as it comes together.",
    "You shuffle the deck slowly, each card sliding past the last with a dry hiss.",
    "You cut the deck three times, then bridge it back together with a satisfying snap.",
    "You shuffle the cards absently, your fingers moving through a motion they know by heart.",
    "You fan the cards out face-down, sweep them back together, and square the deck.",
];

var DRAW_INTROS = [
    "You draw a card from the middle of the deck and turn it over.",
    "You cut the deck and draw from the top.",
    "You draw the bottom card and hold it up to the light.",
    "You spread the deck face-down and let your hand hover until one card seems to call to you.",
    "You draw without looking, then flip the card face-up.",
    "You shuffle once more before drawing, as if giving the deck one last chance to change its mind.",
];

function onCommand_shuffle(user, item, room) {

    var msg = SHUFFLE_MESSAGES[UtilDiceRoll(1, SHUFFLE_MESSAGES.length) - 1];

    SendUserMessage(user.UserId(), msg);
    SendRoomMessage(
        room.RoomId(),
        user.GetCharacterName(true) + " shuffles their <ansi fg=\"item\">" + item.Name(true) + "</ansi>.",
        user.UserId()
    );

    return true;
}

function onCommand_draw(user, item, room) {

    var cardIdx   = UtilDiceRoll(1, MAJOR_ARCANA.length) - 1;
    var card      = MAJOR_ARCANA[cardIdx];
    var reversed  = UtilDiceRoll(1, 2) === 1;
    var intro     = DRAW_INTROS[UtilDiceRoll(1, DRAW_INTROS.length) - 1];

    var cardName  = reversed ? card.name + " <ansi fg=\"red\">(Reversed)</ansi>" : card.name;
    var meaning   = reversed ? card.reversed : card.upright;

    SendUserMessage(
        user.UserId(),
        intro + "\n" +
        "  <ansi fg=\"yellow-bold\">" + cardName + "</ansi>\n" +
        "  <ansi fg=\"white\">" + meaning + ".</ansi>"
    );

    SendRoomMessage(
        room.RoomId(),
        user.GetCharacterName(true) + " draws a card from their <ansi fg=\"item\">" + item.Name(true) + "</ansi>: <ansi fg=\"yellow-bold\">" + cardName + "</ansi>.",
        user.UserId()
    );

    return true;
}
