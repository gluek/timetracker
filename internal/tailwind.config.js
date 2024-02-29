/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./components/**/*.{html,js,templ}"],
  theme: {
    extend: {},
  },
  daisyui: {
    themes: ["dark", "emerald"],
  },
  plugins: [require("daisyui")],
};
