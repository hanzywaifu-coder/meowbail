/**
 * Dongtube-meowbail fluent message builder.
 * Membangun pesan interaktif (native flow) format viewOnceMessage.
 */

class MessageBuilder {
    constructor() {
        this._body = '';
        this._footer = '';
        this._title = '';
        this._headerDoc = null;
        this._headerImg = null;
        this._buttons = [];
        this._contextInfo = {};
    }

    title(t) {
        this._title = t;
        return this;
    }

    body(b) {
        this._body = b;
        return this;
    }

    footer(f) {
        this._footer = f;
        return this;
    }

    document(pathOrBuffer, fileName = 'Document', mimetype = 'image/png') {
        this._headerDoc = {
            mimetype,
            fileName,
            fileLength: 10000,
            pageCount: 100,
            jpegThumbnail: Buffer.isBuffer(pathOrBuffer) ? pathOrBuffer : null
        };
        return this;
    }

    newsletter(jid, name = 'Dongtube Saluran') {
        this._contextInfo = {
            isForwarded: true,
            forwardingScore: 9999,
            forwardedNewsletterMessageInfo: {
                newsletterJid: jid,
                newsletterName: name
            }
        };
        return this;
    }

    addReply(displayText, id) {
        this._buttons.push({
            name: 'quick_reply',
            buttonParamsJson: JSON.stringify({ display_text: displayText, id })
        });
        return this;
    }

    addUrl(displayText, url) {
        this._buttons.push({
            name: 'cta_url',
            buttonParamsJson: JSON.stringify({ display_text: displayText, url, merchant_url: url })
        });
        return this;
    }

    addCopy(displayText, copyCode) {
        this._buttons.push({
            name: 'cta_copy',
            buttonParamsJson: JSON.stringify({ display_text: displayText, copy_code: copyCode })
        });
        return this;
    }

    addSelection(title, sections = []) {
        this._buttons.push({
            name: 'single_select',
            buttonParamsJson: JSON.stringify({ title, sections })
        });
        return this;
    }

    build() {
        const interactiveMessage = {
            header: {
                title: this._title,
                hasMediaAttachment: !!this._headerDoc || !!this._headerImg,
                documentMessage: this._headerDoc
            },
            body: { text: this._body },
            footer: { text: this._footer },
            nativeFlowMessage: {
                buttons: this._buttons
            },
            contextInfo: this._contextInfo
        };

        return {
            viewOnceMessage: {
                message: {
                    interactiveMessage
                }
            }
        };
    }
}

module.exports = { MessageBuilder };
