package din

import (
	"math"
	"math/rand"
	"testing"

	rcmd "github.com/auxten/edgeRec/recommend"
	log "github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	G "gorgonia.org/gorgonia"
	"gorgonia.org/tensor"
)

func TestDin(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	rand.Seed(42)
	var (
		batchSize     = 20
		uProfileDim   = 5
		uBehaviorSize = 3
		uBehaviorDim  = 7
		iFeatureDim   = 7
		cFeatureDim   = 5

		numExamples = 10000
		epochs      = 20
		sampleInfo  = &rcmd.SampleInfo{
			UserProfileRange:  [2]int{0, uProfileDim},
			UserBehaviorRange: [2]int{uProfileDim, uProfileDim + uBehaviorSize*uBehaviorDim},
			ItemFeatureRange:  [2]int{uProfileDim + uBehaviorSize*uBehaviorDim, uProfileDim + uBehaviorSize*uBehaviorDim + iFeatureDim},
			CtxFeatureRange:   [2]int{uProfileDim + uBehaviorSize*uBehaviorDim + iFeatureDim, uProfileDim + uBehaviorSize*uBehaviorDim + iFeatureDim + cFeatureDim},
		}
		inputWidth = uProfileDim + uBehaviorSize*uBehaviorDim + iFeatureDim + cFeatureDim
	)
	inputSlice := make([]float64, numExamples*inputWidth)
	for i := 0; i < numExamples; i++ {
		for j := 0; j < sampleInfo.UserProfileRange[1]; j++ {
			inputSlice[i*inputWidth+j] = rand.Float64()
		}
		for j := sampleInfo.CtxFeatureRange[0]; j < sampleInfo.CtxFeatureRange[1]; j++ {
			inputSlice[i*inputWidth+j] = rand.Float64()
		}
		for j := sampleInfo.UserBehaviorRange[0] + uBehaviorDim; j < sampleInfo.UserBehaviorRange[0]+2*uBehaviorDim; j++ {
			inputSlice[i*inputWidth+j] = rand.Float64()
		}
		for j := sampleInfo.ItemFeatureRange[0]; j < sampleInfo.ItemFeatureRange[1]; j++ {
			inputSlice[i*inputWidth+j] = rand.Float64()
		}
	}
	//for i := 0; i < numExamples*inputWidth; i++ {
	//	inputSlice[i] = rand.Float64()
	//}
	inputs := tensor.New(tensor.WithShape(numExamples, inputWidth), tensor.WithBacking(inputSlice))
	labelSlice := make([]float64, numExamples)
	for i := 0; i < numExamples; i++ {
		//distance of uProfile and cFeature slice
		var dist1, dist2 float64
		for j := 0; j < uProfileDim; j++ {
			dist1 += math.Abs(inputSlice[i*inputWidth+sampleInfo.UserProfileRange[0]+j] - inputSlice[i*inputWidth+sampleInfo.CtxFeatureRange[0]+j])
		}
		labelSlice[i] = dist1 / float64(uProfileDim)

		//distance of 2nd uBehavior and iFeature
		for j := 0; j < uBehaviorDim; j++ {
			dist2 += math.Abs(inputSlice[i*inputWidth+sampleInfo.UserBehaviorRange[0]+uBehaviorDim+j] - inputSlice[i*inputWidth+sampleInfo.ItemFeatureRange[0]+j])
		}
		labelSlice[i] = math.Round((labelSlice[i] + (dist2 / float64(uBehaviorDim) * 0.3)) * 1.1)
	}

	labels := tensor.New(tensor.WithShape(numExamples, 1), tensor.WithBacking(labelSlice))
	log.Debugf("labels: %+v", labels.Data())

	Convey("Din model", t, func() {
		g := G.NewGraph()
		model := NewDinNet(g, uProfileDim, uBehaviorSize, uBehaviorDim, iFeatureDim, cFeatureDim)
		err := Train(uBehaviorSize, uBehaviorDim, uProfileDim, iFeatureDim, cFeatureDim,
			numExamples, batchSize, epochs,
			sampleInfo,
			inputs, labels,
			g,
			model,
		)
		So(err, ShouldBeNil)
	})

	Convey("Simple MLP", t, func() {
		g := G.NewGraph()
		model := NewSimpleMLP(g, uProfileDim, uBehaviorSize, uBehaviorDim, iFeatureDim, cFeatureDim)
		err := Train(uBehaviorSize, uBehaviorDim, uProfileDim, iFeatureDim, cFeatureDim,
			numExamples, batchSize, epochs,
			sampleInfo,
			inputs, labels,
			g,
			model,
		)
		So(err, ShouldBeNil)
	})
}
