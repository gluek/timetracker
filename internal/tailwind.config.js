/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./components/**/*.{html,js,templ}"],
  theme: {
    extend: {},
  },
  daisyui: {
    themes: [
      "dark",
      "emerald",
      {
        pmd_light: {
          primary: "#000000",
          secondary: "#5611EB",
          accent: "#A110EA",
          neutral: "#d1d5db",
          "base-100": "#f3f4f6",
          info: "#0000ff",
          success: "#57EAEA",
          warning: "#F05C00",
          error: "#EA0CEB",
        },
        pmd_dark: {
          primary: "#FFFFFF",
          secondary: "#5611EB",
          accent: "#A110EA",
          neutral: "#d1d5db",
          "base-100": "#000000",
          info: "#4781ED",
          success: "#57EAEA",
          warning: "#F05C00",
          error: "#EA0CEB",
        },
      },
    ],
  },
  plugins: [require("daisyui")],
};
