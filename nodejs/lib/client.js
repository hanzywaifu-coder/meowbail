/**
 * DongtubeMeowbail client (standalone Node.js layer).
 * WebSocket protokol penuh menyusul bridging ke core Go.
 */

const { EventEmitter } = require('events');
const { MessageBuilder } = require('./message-builder');
const { LIDResolver } = require('./lid-resolver');

/**
 * DongtubeMeowbail Client
 */
class DongtubeMeowbail extends EventEmitter {
    constructor(options = {}) {
        super();
        this.options = {
            printQRInTerminal: options.printQRInTerminal ?? true,
            pairingCode: options.pairingCode ?? null,
            phoneNumber: options.phoneNumber ?? null,
            newsletterJid: options.newsletterJid ?? null,
            newsletterName: options.newsletterName ?? null,
            ...options
        };
        this.user = null;
        this.isConnected = false;
        this.lidResolver = new LIDResolver();
        this.retryCounts = new Map();
    }

    createBuilder() {
        return new MessageBuilder();
    }

    async requestPairingCode(phoneNumber, customCode = null) {
        const cleanNumber = (phoneNumber || '').replace(/[^0-9]/g, '');
        if (!cleanNumber) {
            throw new Error('Nomor telepon tidak valid');
        }

        if (customCode) {
            const clean = customCode.replace(/[^a-zA-Z0-9]/g, '').toUpperCase();
            if (clean.length !== 8) {
                throw new Error('Custom pairing code harus tepat 8 karakter!');
            }
            const formatted = `${clean.slice(0, 4)}-${clean.slice(4)}`;
            this.emit('pairing.code', formatted);
            return formatted;
        }

        const chars = '123456789ABCDEFGHJKLMNPQRSTVWXYZ';
        let generated = '';
        for (let i = 0; i < 8; i++) {
            generated += chars[Math.floor(Math.random() * chars.length)];
        }
        const formatted = `${generated.slice(0, 4)}-${generated.slice(4)}`;
        this.emit('pairing.code', formatted);
        return formatted;
    }

    async sendGroupStatus(groupJid, content = {}) {
        if (!groupJid.endsWith('@g.us')) {
            throw new Error('JID harus berupa grup (@g.us)');
        }

        const payload = {
            groupStatusMessageV2: {
                message: content
            }
        };

        return await this.relayMessage(groupJid, payload);
    }

    async sendInteractiveMenu(jid, menuOptions = {}) {
        const builder = new MessageBuilder();
        if (menuOptions.body) builder.body(menuOptions.body);
        if (menuOptions.footer) builder.footer(menuOptions.footer);
        if (menuOptions.thumbnail) builder.document(menuOptions.thumbnail, menuOptions.fileName || 'Menu');
        if (this.options.newsletterJid) builder.newsletter(this.options.newsletterJid, this.options.newsletterName);

        if (menuOptions.sections && menuOptions.sections.length > 0) {
            builder.addSelection('Selection', menuOptions.sections);
        }
        if (menuOptions.ctaUrl) {
            builder.addUrl(menuOptions.ctaText || 'Visit', menuOptions.ctaUrl);
        }
        if (menuOptions.copyCode) {
            builder.addCopy(menuOptions.copyText || 'Copy', menuOptions.copyCode);
        }

        const msg = builder.build();
        return await this.relayMessage(jid, msg);
    }

    async relayMessage(jid, message, options = {}) {
        const messageId = options.messageId || 'MEOWBAIL_' + Date.now();
        this.emit('message.relay', { jid, message, messageId });
        return { key: { remoteJid: jid, id: messageId, fromMe: true }, message };
    }

    async sendCarousel(jid, text, cards = []) {
        const payload = {
            interactiveMessage: {
                body: { text },
                carouselMessage: {
                    cards: cards.map(c => ({
                        header: {
                            title: c.title || '',
                            hasMediaAttachment: !!c.image
                        },
                        body: { text: c.body || '' },
                        footer: { text: c.footer || '' },
                        nativeFlowMessage: {
                            buttons: c.buttons || []
                        }
                    }))
                }
            }
        };
        return await this.relayMessage(jid, { viewOnceMessage: { message: payload } });
    }

    async sendAlbum(jid, items = []) {
        for (const item of items) {
            await this.sendMessage(jid, item);
        }
    }

    async sendWithAntiBan(jid, text, sendFn) {
        // Simulasi presence composing
        this.emit('presence.update', { id: jid, presence: 'composing' });
        // Hitung delay manusiawi
        const delay = Math.min(2500, Math.max(800, (text || '').length * 35 + Math.floor(Math.random() * 200)));
        await new Promise(r => setTimeout(r, delay));
        this.emit('presence.update', { id: jid, presence: 'paused' });
        return await sendFn();
    }

    async sendOrderReview(jid, orderData = {}) {
        const {
            title = 'Konfirmasi Pesanan',
            currency = 'IDR',
            amount = '10000',
            orderId = 'ORD_' + Date.now()
        } = orderData;

        const payload = {
            interactiveMessage: {
                header: { title },
                body: { text: `Total Tagihan: ${currency} ${amount}\nID Pesanan: ${orderId}` },
                footer: { text: 'Dongtube Payment System' },
                nativeFlowMessage: {
                    buttons: [
                        {
                            name: 'review_and_pay',
                            buttonParamsJson: JSON.stringify({
                                type: 'review_and_pay',
                                currency,
                                total_amount: { value: amount, offset: 100 },
                                order_id: orderId
                            })
                        }
                    ]
                }
            }
        };

        return await this.relayMessage(jid, { viewOnceMessage: { message: payload } });
    }

    async sendMessage(jid, content, options = {}) {
        if (content.groupStatusMessage) {
            return await this.sendGroupStatus(jid, content.groupStatusMessage);
        }
        return await this.relayMessage(jid, content, options);
    }
}

function makeWASocket(config = {}) {
    return new DongtubeMeowbail(config);
}

module.exports = { DongtubeMeowbail, makeWASocket };
