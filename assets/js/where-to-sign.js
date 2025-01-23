let currentIndex = 0;
const targets = [
  document.getElementById("petitioner"),
  document.getElementById("consenter-name"),
];
let flashInterval = null;
let flashTimeout = null;
let isFlashing = false;

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
  isFlashing = true;
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

  flashTimeout = setTimeout(() => {
    clearInterval(flashInterval);
    clearBorders();
    isFlashing = false;
  }, 4000);
}
function updateBordersDuringFlashing() {
  if (!isFlashing) return;
  clearBorders();
  let isVisible = true;
  targets.forEach((target) => {
    if (target) {
      const rect = target.getBoundingClientRect();
      drawBorder(rect, isVisible);
    }
  });
}
function whereToSign() {
  clearInterval(flashInterval);
  clearTimeout(flashTimeout);
  clearBorders();

  if (targets.length === 0) return;

  const target = targets[currentIndex];
  if (target) {
    target.scrollIntoView({ behavior: "smooth", block: "center" });
  }

  startFlashingBorders();

  currentIndex = (currentIndex + 1) % targets.length;
}

document.addEventListener("DOMContentLoaded", () => {
  const button = document.getElementById("btn-where-to-sign");
  button.onclick = whereToSign;
  button.style.display = "block";
});

window.addEventListener("scroll", updateBordersDuringFlashing);
window.addEventListener("resize", updateBordersDuringFlashing);
