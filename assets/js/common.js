async function downloadPDF(className, filename, redirectURL) {
	const { jsPDF } = window.jspdf;
	const container = document.querySelector('.'+className);
	const canvas = await html2canvas(container, { scale: 2 });
	const imgData = canvas.toDataURL('image/png');

	const isLandscape = className === 'a4-landscape';
	const orientation = isLandscape ? 'l' : 'p';

	const pdf = new jsPDF(orientation, 'mm', 'a4');
	const pdfWidth = pdf.internal.pageSize.getWidth();
	const pdfHeight = (canvas.height * pdfWidth) / canvas.width;

	pdf.addImage(imgData, 'PNG', 0, 0, pdfWidth, pdfHeight);
	pdf.save(filename+'.pdf');

	window.location.href = redirectURL;
}

async function copyLink() {
	const currentUrl = window.location.href;

	navigator.clipboard.writeText(currentUrl)
		.then(() => {
			const popout = document.getElementById('popout');
			popout.classList.remove('hidden');
			popout.classList.add('visible');

			setTimeout(() => {
				popout.classList.remove('visible');
				popout.classList.add('hidden');
			}, 2000);
		})
		.catch(err => {
			console.error('無法複製網址：', err);
		});
}

async function copyInnerText(containerSelector) {
	const container = document.querySelector(containerSelector);

	if (!container) {
		console.error(`未找到容器: ${containerSelector}`);
		return;
	}

	navigator.clipboard.writeText(container.innerText)
		.then(() => {
			const popout = document.getElementById('popout');
			popout.classList.remove('hidden');
			popout.classList.add('visible');

			setTimeout(() => {
				popout.classList.remove('visible');
				popout.classList.add('hidden');
			}, 2000);
		})
		.catch(err => {
			console.error('無法複製文字', err);
		});
}
