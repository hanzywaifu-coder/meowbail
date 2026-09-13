export interface SectionRow {
    title: string;
    description?: string;
    id: string;
}

export interface Section {
    title: string;
    highlight_label?: string;
    rows: SectionRow[];
}

export interface InteractiveMenuOptions {
    body?: string;
    footer?: string;
    thumbnail?: Buffer | string;
    fileName?: string;
    sections?: Section[];
    ctaUrl?: string;
    ctaText?: string;
    copyCode?: string;
    copyText?: string;
}

export interface SocketConfig {
    printQRInTerminal?: boolean;
    pairingCode?: string;
    phoneNumber?: string;
    newsletterJid?: string;
    newsletterName?: string;
}

export class MessageBuilder {
    title(t: string): this;
    body(b: string): this;
    footer(f: string): this;
    document(pathOrBuffer: Buffer | string, fileName?: string, mimetype?: string): this;
    newsletter(jid: string, name?: string): this;
    addReply(displayText: string, id: string): this;
    addUrl(displayText: string, url: string): this;
    addCopy(displayText: string, copyCode: string): this;
    addSelection(title: string, sections?: Section[]): this;
    build(): any;
}

export interface CarouselCard {
    title?: string;
    body?: string;
    footer?: string;
    image?: Buffer | string;
    buttons?: any[];
}

export interface OrderOptions {
    title?: string;
    currency?: string;
    amount?: string;
    orderId?: string;
}

export class DongtubeMeowbail {
    constructor(options?: SocketConfig);
    createBuilder(): MessageBuilder;
    requestPairingCode(phoneNumber: string, customCode?: string): Promise<string>;
    sendGroupStatus(groupJid: string, content: any): Promise<any>;
    sendInteractiveMenu(jid: string, menuOptions: InteractiveMenuOptions): Promise<any>;
    sendCarousel(jid: string, text: string, cards: CarouselCard[]): Promise<any>;
    sendAlbum(jid: string, items: any[]): Promise<any>;
    sendOrderReview(jid: string, orderData: OrderOptions): Promise<any>;
    sendWithAntiBan(jid: string, text: string, sendFn: () => Promise<any>): Promise<any>;
    sendMessage(jid: string, content: any, options?: any): Promise<any>;
}

export function makeWASocket(config?: SocketConfig): DongtubeMeowbail;
export default makeWASocket;
