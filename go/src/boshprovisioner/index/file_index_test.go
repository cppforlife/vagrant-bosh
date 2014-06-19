package index_test

import (
	boshlog "bosh/logger"
	boshsys "bosh/system"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "boshprovisioner/index"
)

type Key struct {
	Key string
}

type Value struct {
	Name  string
	Count float64
}

type ArrayValue struct{ Names []string }

type StructValue struct{ Name Name }

type Name struct {
	First string
	Last  string
}

var _ = Describe("FileIndex", func() {
	var (
		index FileIndex
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		fs := boshsys.NewOsFileSystem(logger)

		file, err := fs.TempFile("file-index")
		Expect(err).ToNot(HaveOccurred())

		err = fs.RemoveAll(file.Name())
		Expect(err).ToNot(HaveOccurred())

		index = NewFileIndex(file.Name(), fs)
	})

	Describe("Save/List", func() {
		It("returns list of saved items", func() {
			k1 := Key{Key: "key-1"}
			v1 := Value{Name: "value-1", Count: 1}
			err := index.Save(k1, v1)
			Expect(err).ToNot(HaveOccurred())

			k2 := Key{Key: "key-2"}
			v2 := Value{Name: "value-2", Count: 2}
			err = index.Save(k2, v2)
			Expect(err).ToNot(HaveOccurred())

			var values []Value

			err = index.List(&values)
			Expect(err).ToNot(HaveOccurred())
			Expect(values).To(Equal([]Value{v1, v2}))
		})

		Describe("array values", func() {
			It("returns list of saved items that have nil, 0, 1, and more items in an array", func() {
				k1 := Key{Key: "key-1"}
				v1 := ArrayValue{Names: []string{"name-1-1", "name-1-2"}} // multiple
				err := index.Save(k1, v1)
				Expect(err).ToNot(HaveOccurred())

				k2 := Key{Key: "key-2"}
				v2 := ArrayValue{Names: []string{"name-2-1"}} // single
				err = index.Save(k2, v2)
				Expect(err).ToNot(HaveOccurred())

				k3 := Key{Key: "key-3"}
				v3 := ArrayValue{Names: []string{}} // empty slice
				err = index.Save(k3, v3)
				Expect(err).ToNot(HaveOccurred())

				k4 := Key{Key: "key-4"}
				v4 := ArrayValue{} // nil
				err = index.Save(k4, v4)
				Expect(err).ToNot(HaveOccurred())

				var values []ArrayValue

				err = index.List(&values)
				Expect(err).ToNot(HaveOccurred())
				Expect(values).To(Equal([]ArrayValue{v1, v2, v3, v4}))
			})
		})

		Describe("struct values", func() {
			It("returns list of saved items that have nil or more item", func() {
				k1 := Key{Key: "key-1"}
				v1 := StructValue{Name: Name{First: "first-name-1", Last: "last-name-1"}} // struct
				err := index.Save(k1, v1)
				Expect(err).ToNot(HaveOccurred())

				k2 := Key{Key: "key-2"}
				v2 := StructValue{Name: Name{First: "first-name-1"}} // struct incomplete
				err = index.Save(k2, v2)
				Expect(err).ToNot(HaveOccurred())

				k3 := Key{Key: "key-3"}
				v3 := StructValue{} // zero value
				err = index.Save(k3, v3)
				Expect(err).ToNot(HaveOccurred())

				var values []StructValue

				err = index.List(&values)
				Expect(err).ToNot(HaveOccurred())
				Expect(values).To(Equal([]StructValue{v1, v2, v3}))
			})
		})
	})

	Describe("Save/ListKeys", func() {
		It("returns list of saved keys", func() {
			k1 := Key{Key: "key-1"}
			v1 := Value{Name: "value-1", Count: 1}
			err := index.Save(k1, v1)
			Expect(err).ToNot(HaveOccurred())

			k2 := Key{Key: "key-2"}
			v2 := Value{Name: "value-2", Count: 2}
			err = index.Save(k2, v2)
			Expect(err).ToNot(HaveOccurred())

			var keys []Key

			err = index.ListKeys(&keys)
			Expect(err).ToNot(HaveOccurred())
			Expect(keys).To(Equal([]Key{k1, k2}))
		})
	})

	Describe("Save/Find", func() {
		It("returns true if item is found by key", func() {
			k1 := Key{Key: "key-1"}
			v1 := Value{Name: "value-1", Count: 1}
			err := index.Save(k1, v1)
			Expect(err).ToNot(HaveOccurred())

			var value Value

			err = index.Find(k1, &value)
			Expect(err).ToNot(HaveOccurred())
			Expect(err).ToNot(Equal(ErrNotFound))

			Expect(value).To(Equal(v1))
		})

		It("returns false if item is not found by key", func() {
			k1 := Key{Key: "key-1"}
			v1 := Value{Name: "value-1", Count: 1}
			err := index.Save(k1, v1)
			Expect(err).ToNot(HaveOccurred())

			var value Value

			err = index.Find(Key{Key: "key-2"}, &value)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(ErrNotFound))

			Expect(value).To(Equal(Value{}))
		})

		Describe("array values", func() {
			It("returns true and correctly deserializes item with nil", func() {
				k1 := Key{Key: "key-1"}
				v1 := ArrayValue{} // nil
				err := index.Save(k1, v1)
				Expect(err).ToNot(HaveOccurred())

				var value ArrayValue

				err = index.Find(k1, &value)
				Expect(err).ToNot(HaveOccurred())
				Expect(err).ToNot(Equal(ErrNotFound))

				Expect(value).To(Equal(v1))
			})

			It("returns true and correctly deserializes item with empty slice", func() {
				k1 := Key{Key: "key-1"}
				v1 := ArrayValue{Names: []string{}} // empty slice
				err := index.Save(k1, v1)
				Expect(err).ToNot(HaveOccurred())

				var value ArrayValue

				err = index.Find(k1, &value)
				Expect(err).ToNot(HaveOccurred())
				Expect(err).ToNot(Equal(ErrNotFound))

				Expect(value).To(Equal(v1))
			})

			It("returns true and correctly deserializes item with multiple items", func() {
				k1 := Key{Key: "key-1"}
				v1 := ArrayValue{Names: []string{"name-1-1", "name-1-2"}} // multiple
				err := index.Save(k1, v1)
				Expect(err).ToNot(HaveOccurred())

				var value ArrayValue

				err = index.Find(k1, &value)
				Expect(err).ToNot(HaveOccurred())
				Expect(err).ToNot(Equal(ErrNotFound))

				Expect(value).To(Equal(v1))
			})
		})

		Describe("struct values", func() {
			It("returns true and correctly deserializes item with zero value", func() {
				k1 := Key{Key: "key-1"}
				v1 := StructValue{} // zero value
				err := index.Save(k1, v1)
				Expect(err).ToNot(HaveOccurred())

				var value StructValue

				err = index.Find(k1, &value)
				Expect(err).ToNot(HaveOccurred())
				Expect(err).ToNot(Equal(ErrNotFound))

				Expect(value).To(Equal(v1))
			})

			It("returns true and correctly deserializes item with filled struct", func() {
				k1 := Key{Key: "key-1"}
				v1 := StructValue{Name: Name{First: "first-name-1", Last: "last-name-1"}} // struct
				err := index.Save(k1, v1)
				Expect(err).ToNot(HaveOccurred())

				var value StructValue

				err = index.Find(k1, &value)
				Expect(err).ToNot(HaveOccurred())
				Expect(err).ToNot(Equal(ErrNotFound))

				Expect(value).To(Equal(v1))
			})
		})
	})

	Describe("Save/Remove", func() {
		var (
			k1 Key
			v1 Value
		)

		BeforeEach(func() {
			k1 = Key{Key: "key-1"}
			v1 = Value{Name: "value-1", Count: 1}
			err := index.Save(k1, v1)
			Expect(err).ToNot(HaveOccurred())
		})

		It("removes matching value if found", func() {
			k2 := Key{Key: "key-2"}
			v2 := Value{Name: "value-2", Count: 2}
			err := index.Save(k2, v2)
			Expect(err).ToNot(HaveOccurred())

			err = index.Remove(k1)
			Expect(err).ToNot(HaveOccurred())

			var values []Value

			err = index.List(&values)
			Expect(err).ToNot(HaveOccurred())
			Expect(values).To(Equal([]Value{v2}))
		})

		It("does not remove non-matching value if not found", func() {
			k2 := Key{Key: "key-2"}

			err := index.Remove(k2)
			Expect(err).ToNot(HaveOccurred())

			var values []Value

			err = index.List(&values)
			Expect(err).ToNot(HaveOccurred())
			Expect(values).To(Equal([]Value{v1}))
		})
	})
})
