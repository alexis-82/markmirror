export interface EmojiCategory {
    name: string;
    icon: string;
    emojis: string[];
}

// Emoji separate da spazi. Alcune contengono caratteri invisibili (selettori di
// variante, giunzioni): vanno copiate intere, non ridigitate.
const category = (name: string, icon: string, emojis: string): EmojiCategory => ({name, icon, emojis: emojis.split(' ')});

export const EMOJI_CATEGORIES: EmojiCategory[] = [
    category('Faccine', '😀', '😀 😃 😄 😁 😆 😅 🤣 😂 🙂 🙃 😉 😊 😇 🥰 😍 🤩 😘 😋 😜 🤪 🤑 🤗 🤔 🤐 🤨 😐 😑 😶 😏 😒 🙄 😬 😌 😔 😪 😴 😷 🤒 🤕 🤢 🤮 🥵 🥶 🤯 🥳 😎 🤓 🧐 😕 😟 🙁 😮 😲 😳 🥺 😨 😰 😢 😭 😱 😩 😤 😡 🤬 😈 💀 💩 🤡 👻 👽'),
    category('Persone e gesti', '👍', '👍 👎 👌 ✌️ 🤞 🤟 🤘 🤙 👈 👉 👆 👇 ☝️ ✋ 👋 🤚 🖐️ 👏 🙌 🤝 🙏 💪 ✊ 👊 ✍️ 🧠 👀 👁️ 👤 👥 🧑‍💻 🙋 🤷 🤦'),
    category('Cuori', '❤️', '❤️ 🧡 💛 💚 💙 💜 🖤 🤍 🤎 💔 💕 💖 💯 💢 💫 💬 💭'),
    category('Stato', '✅', '✅ ☑️ ✔️ ❌ ❎ ⚠️ 🚫 ⛔ 🚨 ❗ ❓ ‼️ ⁉️ ℹ️ 🆕 🆗 🆙 🔝 🔜 🔙 ➡️ ⬅️ ⬆️ ⬇️ ↗️ ↘️ 🔄 🔁 ➕ ➖ ♻️ 🔴 🟠 🟡 🟢 🔵 🟣 ⚫ ⚪ 🟥 🟧 🟨 🟩 🟦 ⭐ 🌟 ✨ 🎯 🏆 🥇 🥈 🥉 🎉 🎊 🎁'),
    category('Documenti', '📝', '📌 📍 🔖 🏷️ 📝 ✏️ 🖊️ 📎 🔗 📄 📃 📑 📋 📁 📂 🗂️ 📚 📖 📘 📙 📗 📕 📰 📅 📆 🗓️ ⏰ ⏱️ ⏳ ⌛ 🔍 🔎 🔔 🔕 📢 📣 💡 ✉️ 📧 📨 📤 📥'),
    category('Tecnologia', '💻', '💻 🖥️ ⌨️ 🖱️ 📱 🔌 🔋 💾 💿 🖨️ 📷 🎥 📺 📡 🌐 🔒 🔓 🔐 🔑 🛡️ ⚙️ 🔧 🔨 🛠️ 🧰 🧪 🧬 🔬 🐛 🚀 📦 🏗️ 🚧 🧹 🗑️ 🔥 ⚡ 💥 🤖 📊 📈 📉 🧩 💰 💳'),
    category('Vari', '🌍', '☀️ 🌙 ☁️ 🌧️ ❄️ 🌈 🌍 🌱 🌳 🌸 🍀 🐶 🐱 🐧 🦄 🐝 ☕ 🍺 🍕 🍎 🍰 🏠 🏢 🚗 ✈️ 🚂 🗺️ ⚽ 🎮 🎵 🎬 🎨'),
    category('Bandiere', '🏁', '🏁 🚩 🇮🇹 🇬🇧 🇺🇸 🇪🇺 🇫🇷 🇩🇪 🇪🇸 🇵🇹 🇯🇵 🇨🇳'),
];
