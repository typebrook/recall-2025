const targets = [
  document.getElementById("petitioner"),
  document.getElementById("consenter-name"),
];
let flashInterval = null;

function drawBorder(rect, isVisible) {
  const overlayContainer = document.getElementById("where-to-sign");
  const borderDiv = document.createElement("div");
  borderDiv.style.position = "absolute";
  borderDiv.style.border = isVisible ? "4px dashed #ff073a" : "none";
  borderDiv.style.pointerEvents = "none";
  borderDiv.style.top = `${rect.top}px`;
  borderDiv.style.left = `${rect.left}px`;
  borderDiv.style.width = `${rect.width}px`;
  borderDiv.style.height = `${rect.height}px`;
  overlayContainer.appendChild(borderDiv);
}

function clearBorders() {
  const overlayContainer = document.getElementById("where-to-sign");
  overlayContainer.innerHTML = "";
}

function startFlashingBorders() {
  clearBorders();
  let isVisible = true;

  flashInterval = setInterval(() => {
    clearBorders();
    targets.forEach((target) => {
      if (target) {
        const rect = target.getBoundingClientRect();
        drawBorder(rect, isVisible);
      }
    });
    isVisible = !isVisible;
  }, 500);
}

document.addEventListener("DOMContentLoaded", () => {
  startFlashingBorders();
});

window.addEventListener("scroll", () => {
  clearBorders();
  targets.forEach((target) => {
    if (target) {
      const rect = target.getBoundingClientRect();
      drawBorder(rect, true);
    }
  });
});
