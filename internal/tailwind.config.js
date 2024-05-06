/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./components/**/*.{html,js,templ}"],
  theme: {
    extend: {},
  },
  plugins: [require("daisyui")],
  daisyui: {
    themes: ["emerald", "dark"],
    darkTheme: "dark",
  },
};
