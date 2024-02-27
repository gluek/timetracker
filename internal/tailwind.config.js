/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./components/**/*.{html,js,templ}"],
  theme: {
    extend: {},
  },
  daisyui: {
    themes: ["light", "dark", "cupcake", "night"],
  },
  plugins: [require("daisyui")],
};
