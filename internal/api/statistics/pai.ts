import { EChartsOption } from "echarts"
import { filterData, renderChart, StatisticData } from "./statistics";

interface PaiData extends StatisticData {
	pai: number
}

export function InitPaiGraph(id: string, title: string, data: PaiData[]) {
	// @ts-expect-error Declared globally
	OnElementReady("#" + id, () => {
		createPaiGraph(id, title, filterData(data))
	})
}

function createPaiGraph(id: string, title: string, data: PaiData[]) {
	const options: EChartsOption = {
		title: {
			text: title,
			textAlign: "center"
		},
		xAxis: [
			{
				type: "category",
				data: data.map(d => d.label)
			}
		],
		yAxis: [
			{
				type: "value"
				
			}
		],
		series: {
			type: "bar",
			data: data.map(d => d.pai),
		}
	}

	renderChart(id, true, options)
}