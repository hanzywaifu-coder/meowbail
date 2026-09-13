/**
 * LIDResolver: memetakan LID <-> nomor asli (PN) untuk fitur tag/mention.
 */

class LIDResolver {
    constructor() {
        this.lidToPn = new Map();
        this.pnToLid = new Map();
    }

    registerMapping(lid, pn) {
        if (!lid || !pn) return;
        this.lidToPn.set(lid.toLowerCase(), pn.toLowerCase());
        this.pnToLid.set(pn.toLowerCase(), lid.toLowerCase());
    }

    resolveToPN(jid) {
        if (!jid) return jid;
        return this.lidToPn.get(jid.toLowerCase()) || jid;
    }

    resolveToLID(jid) {
        if (!jid) return jid;
        return this.pnToLid.get(jid.toLowerCase()) || jid;
    }
}

module.exports = { LIDResolver };
