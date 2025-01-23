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

  if (isVisible) {
    const textDiv = document.createElement("div");
    textDiv.textContent = "在這簽名";
    textDiv.style.position = "absolute";
    textDiv.style.top = "50%";
    textDiv.style.left = "50%";
    textDiv.style.transform = "translate(-50%, -50%)";
    textDiv.style.color = "#ff073a";
    textDiv.style.fontSize = "18px";
    textDiv.style.fontWeight = "bold";
    textDiv.style.pointerEvents = "none";

    borderDiv.appendChild(textDiv);
  }

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
