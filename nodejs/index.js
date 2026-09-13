/**
 * dongtube-meowbail — Node.js layer (standalone).
 *
 * Entry point yang merangkum modul-modul kecil di lib/.
 * Layer ini saat ini berdiri sendiri; bridging ke core Go menyusul
 * setelah core stabil (lihat REBUILD.md di root repo).
 */

const { MessageBuilder } = require('./lib/message-builder');
const { LIDResolver } = require('./lib/lid-resolver');
const { DongtubeMeowbail, makeWASocket } = require('./lib/client');

module.exports = {
    makeWASocket,
    DongtubeMeowbail,
    MessageBuilder,
    LIDResolver,
    default: makeWASocket
};
